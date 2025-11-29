package server

// TODO.
type Thread struct {
	ID    int    `json:"ID,omitempty"`
	Title string `json:"Title"`
	Body  string `json:"Body"`
}
