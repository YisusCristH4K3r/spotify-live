package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type webAccessTokenResponse struct {
	ClientId                         string `json:"clientId"`
	AccessToken                      string `json:"accessToken"`
	AccessTokenExpirationTimestampMs int64  `json:"accessTokenExpirationTimestampMs"`
	IsAnonymous                      bool   `json:"isAnonymous"`
}

type ApiClient struct {
	spDcCookie    string
	token         *webAccessTokenResponse
	tokenMutex    sync.Mutex
	tokenCond     *sync.Cond
	refreshing    bool
	refreshNotify chan struct{}
}

func NewApiClient(spDcCookie string) *ApiClient {
	client := &ApiClient{spDcCookie: spDcCookie, refreshNotify: make(chan struct{})}
	client.tokenCond = sync.NewCond(&client.tokenMutex)
	return client
}

// Do sends an HTTP request, refreshing the token if necessary.
func (api *ApiClient) Do(req *http.Request) (*http.Response, error) {
	if err := api.ensureTokenValid(); err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+api.token.AccessToken)
	client := &http.Client{Timeout: 70 * time.Second}
	return client.Do(req)
}

// ensureTokenValid refreshes the token if it's expired.
func (api *ApiClient) ensureTokenValid() error {
	api.tokenMutex.Lock()
	defer api.tokenMutex.Unlock()

	// Wait while a refresh is already in progress
	for api.refreshing {
		api.tokenCond.Wait()
	}

	// Check if token is still valid
	if api.token == nil || time.Now().After(time.UnixMilli(api.token.AccessTokenExpirationTimestampMs)) {
		api.refreshing = true
		api.tokenMutex.Unlock() // Unlock while refreshing the token

		err := api.refreshToken()

		api.tokenMutex.Lock() // Re-lock to update state
		api.refreshing = false
		api.tokenCond.Broadcast() // Notify all waiting goroutines

		if err != nil {
			return err
		}
	}
	return nil
}

// refreshToken refreshes the access token using the refresh token.
func (api *ApiClient) refreshToken() error {
	fmt.Println("Refreshing token...")
	url := "https://open.spotify.com/get_access_token?reason=transport&productType=web_player"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Set the cookies in the header
	req.Header.Add("Cookie", "sp_dc="+api.spDcCookie)

	// Create an HTTP client with a timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tokenResponse webAccessTokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return err
	}

	api.token = &tokenResponse
	return nil
}

func (api *ApiClient) GetFriendActivity() (*FriendActivityResponse, error) {
	url := "https://guc-spclient.spotify.com/presence-view/v1/buddylist"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Send the request
	resp, err := api.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var activityResponse FriendActivityResponse
	err = json.Unmarshal(body, &activityResponse)
	if err != nil {
		fmt.Printf("\nError parsing response: %s", body)
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		return nil, err
	}

	return &activityResponse, nil
}
