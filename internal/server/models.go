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
	Bio          string    `json:"bio,omitempty" db:"bio"`
	Avatar       int       `json:"avatar,omitempty" db:"avatar"`
	IsAdmin      bool      `json:"-" db:"is_admin"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Thread is a struct that holds all the data needed for thread functionality.
type Thread struct {
	ID         int       `json:"thread_id,omitempty" db:"thread_id"`
	Title      string    `json:"title" db:"title"`
	Body       string    `json:"body" db:"body"`
	AuthorID   int       `json:"author_id,omitempty" db:"author_id"`
	CategoryID int       `json:"category_id,omitempty" db:"category_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Category is a struct for managing categories in the app.
type Category struct {
	ID          int       `json:"category_id,omitempty" db:"category_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	AuthorID    int       `json:"author_id,omitempty" db:"author_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// A Comment is a post responding to a thread.
type Comment struct {
	ID        int       `json:"comment_id,omitempty" db:"comment_id"`
	ThreadID  int       `json:"thread_id,omitempty" db:"thread_id"`
	AuthorID  int       `json:"author_id,omitempty" db:"author_id"`
	Body      string    `json:"body" db:"body"`
	ReplyID   int       `json:"reply_id,omitempty" db:"reply_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ThreadView is the view model for a thread as shown in list pages (home, category).
type ThreadView struct {
	CategoryID      int       `db:"category_id"`
	ThreadID        int       `db:"thread_id"`
	Title           string    `db:"title"`
	AuthorName      string    `db:"username"`
	AuthorID        int       `db:"author_id"`
	Pinned          bool      `db:"pinned"`
	ReplyCount      int       `db:"reply_count"`
	LatestComment   string    `db:"latest_comment"`
	LatestCommentID int       `db:"latest_comment_id"`
	LatestTS        time.Time `db:"latest_ts"`
}

// ThreadDetail is the full view model for a single thread page.
type ThreadDetail struct {
	Title         string    `db:"title"`
	ThreadID      int       `db:"thread_id"`
	Body          string    `db:"body"`
	AuthorID      int       `db:"author_id"`
	Avatar        int       `db:"avatar"`
	AvatarStr     string    `db:"-"`
	Username      string    `db:"username"`
	CategoryID    int       `db:"category_id"`
	CategoryTitle string    `db:"category_title"`
	CreatedAt     time.Time `db:"created_at"`
}

// CommentView is the view model for a comment as shown on a thread page.
type CommentView struct {
	AuthorID            int       `db:"author_id"`
	Avatar              int       `db:"avatar"`
	AvatarStr           string    `db:"-"`
	Username            string    `db:"username"`
	CommentID           int       `db:"comment_id"`
	ReplyID             int       `db:"reply_id"`
	ReplyPage           int       `db:"reply_page"`
	ReplyBody           string    `db:"reply_body"`
	ReplyAuthorUsername string    `db:"reply_author_username"`
	ReplyAuthorID       int       `db:"reply_author_id"`
	Body                string    `db:"body"`
	CreatedAt           time.Time `db:"created_at"`
}

// GeneratePasswordHash adds a hashed password to a User struct if  there is a
// password in the struct, and a password hash is not already present.
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
