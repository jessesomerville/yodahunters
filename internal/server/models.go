package server

// TODO.

// Thread is a simplified struct for testing thread functionality.
type Thread struct {
	ID    int    `json:"ID,omitempty"`
	Title string `json:"Title"`
	Body  string `json:"Body"`
}
