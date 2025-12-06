package server

// TODO.
import (
	"time"
)

// User is a struct for managing users in the app.
// PasswordHash is omitted completely from JSON since it should never be returned.
// Password can be supplied on registration, but it won't be saved in the DB so
// we shouldn't have to worry about it getting marshaled accidentally.
type User struct {
	ID           int       `json:"id,omitempty" db:"id"`
	Username     string    `json:"username,omitempty" db:"username"`
	Email        string    `json:"email,omitempty" db:"email"`
	Password     string    `json:"password,omitempty"`
	PasswordHash []byte    `json:"-" db:"pw_hash"`
	CreatedAt    time.Time `json:"created_at,omitempty" db:"created_at"`
}

// Thread is a simplified struct for testing thread functionality.
type Thread struct {
	ID        int       `json:"id,omitempty" db:"id"`
	Title     string    `json:"Title" db:"title"`
	Body      string    `json:"Body" db:"body"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
}
