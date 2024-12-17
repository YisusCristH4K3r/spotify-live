package spotify

type User struct {
	Uri      string `json:"uri"`
	Name     string `json:"name"`
	ImageUrl string `json:"imageUrl,omitempty"`
}

type Album struct {
	Uri  string `json:"uri"`
	Name string `json:"name"`
}

type Artist struct {
	Uri  string `json:"uri"`
	Name string `json:"name"`
}

type TrackContext struct {
	Uri   string `json:"uri"`
	Name  string `json:"name"`
	Index int    `json:"index"`
}

type Track struct {
	Uri      string       `json:"uri"`
	Name     string       `json:"name"`
	ImageUrl string       `json:"imageUrl"`
	Album    Album        `json:"album"`
	Artist   Artist       `json:"artist"`
	Context  TrackContext `json:"context"`
}

type Friend struct {
	Timestamp int64 `json:"timestamp"`
	User      User  `json:"user"`
	Track     Track `json:"track"`
}

type FriendActivityResponse struct {
	Friends []Friend `json:"friends"`
}
