package spotify

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
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
	spotifySecret spotifySecret
}

type spotifySecret struct {
	Version int       `json:"version"`
	Secret  string    `json:"secret"`
	Updated time.Time `json:"updated,omitempty"`
}

func NewApiClient(spDcCookie string) *ApiClient {
	client := &ApiClient{spDcCookie: spDcCookie, refreshNotify: make(chan struct{})}
	client.tokenCond = sync.NewCond(&client.tokenMutex)
	return client
}

// Do send an HTTP request, refreshing the token if necessary.
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

	totpCode, err := api.generateTotp()
	if err != nil {
		return err
	}
	timestamp := time.Now().Unix()
	baseURL := "https://open.spotify.com/api/token"

	// Convert serverTime to UTC and format as "YYYY-MM-DD"
	serverTime := time.Now().Unix()
	buildDate := time.Unix(serverTime, 0).UTC().Format("2006-01-02")

	// Generate a random 4-byte hexadecimal string
	randomBytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err = rand.Read(randomBytes)
	if err != nil {
		panic(err) // Handle error in production code
	}
	randomHex := hex.EncodeToString(randomBytes)

	// Create the buildVer string
	buildVer := fmt.Sprintf("web-player_%s_%d_%s", buildDate, serverTime*1000, randomHex)

	// Define query parameters
	params := url.Values{}
	params.Add("reason", "transport")
	params.Add("productType", "web-player")
	params.Add("totp", totpCode)
	params.Add("totpServer", totpCode)
	params.Add("totpVer", strconv.FormatInt(int64(api.spotifySecret.Version), 10))
	params.Add("sTime", strconv.FormatInt(serverTime, 10))
	params.Add("cTime", strconv.FormatInt(timestamp, 10))
	params.Add("buildDate", buildDate)
	params.Add("buildVer", buildVer)

	// Construct the full URL
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}

	// Set the cookies in the header
	req.Header.Add("Cookie", "sp_dc="+api.spDcCookie)
	//req.Header.Add("User-Agent",USER_AGENT)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Referer", "https://open.spotify.com/")
	req.Header.Add("App-Platform", "WebPlayer")

	// Create an HTTP client with a timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Spotify refresh token API returned status code %d", resp.StatusCode)
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

func (api *ApiClient) refreshSecrets() error {
	req, err := http.NewRequest("GET", "https://github.com/xyloflake/spot-secrets-go/blob/main/secrets/secrets.json?raw=true", nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second}
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
	var secrets []spotifySecret
	err = json.Unmarshal(body, &secrets)
	if err != nil {
		return err
	}

	api.spotifySecret = secrets[len(secrets)-1]
	api.spotifySecret.Updated = time.Now()
	return nil
}

type ServerTimeResponse struct {
	ServerTime int64 `json:"serverTime"` // Assuming server time is a Unix timestamp (seconds)
}

// generateTotp generates the secret, fetches server time, and computes the TOTP.
// Returns the TOTP code and any error encountered.
func (api *ApiClient) generateTotp() (string, error) {
	// Refetch secrets once a day
	if api.spotifySecret.Updated.Before(time.Now().Add(24 * time.Hour)) {
		err := api.refreshSecrets()
		if err != nil {
			return "", err
		}
	}

	// 1. Secret Generation Logic (mirrors the Python code)
	secretCipherBytes := []byte(api.spotifySecret.Secret)

	transformed := make([]byte, len(secretCipherBytes))
	transformedStr := make([]string, len(secretCipherBytes))
	for t, e := range secretCipherBytes {
		// Perform the XOR transformation
		transformed[t] = byte(int(e) ^ ((t % 33) + 9))
		// Convert the transformed byte to its string representation
		transformedStr[t] = strconv.Itoa(int(transformed[t]))
	}

	// Join the string representations
	joined := strings.Join(transformedStr, "")

	// Encode the joined string to UTF-8 bytes (already bytes in Go)
	utf8Bytes := []byte(joined)

	// Convert UTF-8 bytes to a hex string
	hexStr := hex.EncodeToString(utf8Bytes)

	// Convert the hex string back to bytes
	secretBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex string: %w", err)
	}

	// Base32 encode the resulting bytes, removing padding
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)

	// Configure HTTP client with timeout
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Create the HTTP request
	req, err := http.NewRequest("HEAD", "https://open.spotify.com/", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Host", "open.spotify.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:137.0) Gecko/20100101 Firefox/137.0")
	req.Header.Set("Accept", "*/*")

	// Execute the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close() // Ensure response body is closed

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	// Decode JSON response
	serverDate := resp.Header.Get("Date")
	if serverDate == "" {
		return "", fmt.Errorf("failed to get server date from response")
	}
	serverTime, err := time.Parse(time.RFC1123, serverDate)
	if err != nil {
		return "", fmt.Errorf("failed to parse server date: %w", err)
	}

	// 3. TOTP Generation
	// Generate the TOTP code using the fetched server time
	// The pquerna library uses time.Now() by default if opts is nil,
	// so we must provide the specific time.
	opts := totp.ValidateOpts{
		Period:    30,                // Standard TOTP interval
		Digits:    6,                 // Standard TOTP digits
		Algorithm: otp.AlgorithmSHA1, // Standard TOTP algorithm
	}
	otpCode, err := totp.GenerateCodeCustom(secret, serverTime, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP code: %w", err)
	}

	return otpCode, nil
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
