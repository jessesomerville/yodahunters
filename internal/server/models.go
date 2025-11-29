package server

// TODO.
import (
	"time"
)

// User is a struct for managing users in the app.
type User struct {
	ID        int       `json:"id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// Thread is a simplified struct for testing thread functionality.
type Thread struct {
	ID        int       `json:"id,omitempty"`
	Title     string    `json:"Title"`
	Body      string    `json:"Body"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}
