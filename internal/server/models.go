package server

// TODO.
import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
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

func (u *User) GeneratePasswordHash() error {
	// If no password is provided, then it should be set to an empty string.
	// It shouldn't be possible to set an empty string as a password in any case.
	if u.PasswordHash != nil {
		return errors.New("user struct already contains password hash")
	}
	if u.Password == "" {
		return errors.New("attempted to generate password hash where password is an empty string")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	u.Password = ""
	return nil
}

// Thread is a simplified struct for testing thread functionality.
type Thread struct {
	ID        int       `json:"id,omitempty" db:"id"`
	Title     string    `json:"Title" db:"title"`
	Body      string    `json:"Body" db:"body"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
}
