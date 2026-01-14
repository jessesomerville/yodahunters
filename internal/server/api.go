package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jessesomerville/yodahunters/internal/pg"
	"github.com/jessesomerville/yodahunters/internal/server/middleware"
	"golang.org/x/crypto/bcrypt"
)

// This is ugly but I think it is one of the faster ways to do, and a lot
// of requests are going to hit it.
func pageBuilder(q string, r *http.Request) string {
	var sb strings.Builder
	var page middleware.Page
	page = r.Context().Value(middleware.CtxPageKey).(middleware.Page)
	offset := strconv.Itoa(page.Size * (page.Number - 1))
	size := strconv.Itoa(page.Size)
	sb.WriteString(q)
	sb.WriteString(" ORDER BY created_at DESC")
	sb.WriteString(" OFFSET ")
	sb.WriteString(offset)
	sb.WriteString(" LIMIT ")
	sb.WriteString(size)
	return sb.String()
}

func (s *Server) apiHandleGetThreads(w http.ResponseWriter, r *http.Request) error {
	q := pageBuilder(`SELECT thread_id, author_id, category_id, title, body, created_at FROM threads`, r)
	threads, err := pg.QueryRowsToStruct[Thread](r.Context(), s.dbClient, q)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(threads)
}

func (s *Server) getHandleGetThreadsByCategoryID(w http.ResponseWriter, r *http.Request) error {
	q := pageBuilder(`SELECT thread_id, author_id, category_id, title, body, created_at FROM threads WHERE category_id = $1`, r)
	categoryID := r.PathValue("id")
	threads, err := pg.QueryRowsToStruct[Thread](r.Context(), s.dbClient, q, categoryID)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(threads)
}

func (s *Server) apiHandleGetThreadByID(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid thread ID %q", r.PathValue("id"))
	}

	q := pageBuilder(`SELECT thread_id, author_id, title, body, created_at FROM threads WHERE id = $1`, r)
	thread, err := pg.QueryRowToStruct[Thread](r.Context(), s.dbClient, q, id)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(thread)
}

func (s *Server) apiHandlePostThreads(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var t Thread
	if err := json.Unmarshal(reqBody, &t); err != nil {
		return err
	}

	const q = `
	INSERT INTO threads (title, body, category_id, author_id)
	VALUES ($1, $2, $3, $4)
	RETURNING thread_id, author_id, category_id, title, body, created_at`

	thread, err := pg.QueryRowToStruct[Thread](r.Context(), s.dbClient, q, t.Title, t.Body, t.CategoryID, r.Context().Value(middleware.CtxUserKey))
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(thread)
}

func (s *Server) apiHandleRegister(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var u User
	if err := json.Unmarshal(reqBody, &u); err != nil {
		return err
	}

	const checkUserExists = "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	var userExists bool
	row, err := s.dbClient.QueryRow(r.Context(), checkUserExists, u.Username)
	if err != nil {
		return err
	}
	row.Scan(&userExists)
	if userExists {
		return fmt.Errorf("user with username: %s already exists", u.Username)
	}

	const checkEmailExists = "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	var emailExists bool
	row, err = s.dbClient.QueryRow(r.Context(), checkEmailExists, u.Email)
	if err != nil {
		return err
	}
	row.Scan(&emailExists)
	if emailExists {
		return fmt.Errorf("user with username: %s already exists", u.Username)
	}

	if err = u.GeneratePasswordHash(); err != nil {
		return err
	}
	const insertUser = `
	INSERT INTO users (username, email, pw_hash, bio, avatar)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, username, email, bio, avatar, created_at`
	row, err = s.dbClient.QueryRow(r.Context(), insertUser, u.Username, u.Email, u.PasswordHash, u.Bio, u.Avatar)
	if err != nil {
		return err
	}
	var user User
	row.Scan(&user.ID, &user.Username, &user.Email, &user.Bio, &user.Avatar, &user.CreatedAt)
	return json.NewEncoder(w).Encode(user)
}

func (s *Server) apiHandleLogin(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var login struct {
		Username string
		Password string
	}
	if err := json.Unmarshal(reqBody, &login); err != nil {
		return err
	}

	const q = "SELECT id, pw_hash, is_admin FROM users WHERE username = $1"
	row, err := s.dbClient.QueryRow(r.Context(), q, login.Username)
	if err != nil {
		return err
	}
	var id int
	var passwordHash []byte
	var isAdmin bool
	if err = row.Scan(&id, &passwordHash, &isAdmin); err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(login.Password))
	if err != nil {
		return err
	}

	jwt, err := middleware.GenerateJWT(id, isAdmin, s.jwtSecret)
	if err != nil {
		return err
	}
	var token struct {
		AccessToken string `json:"access_token"`
	}
	token.AccessToken = jwt.Raw

	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    jwt.Raw,
		Expires:  time.Now().Add(12 * time.Hour), // Set an expiration time
		Path:     "/",                            // Make the cookie available to all paths
		HttpOnly: true,
		// Secure: true,
	}

	http.SetCookie(w, cookie)

	return json.NewEncoder(w).Encode(token)
}

func (s *Server) apiHandleGetMe(w http.ResponseWriter, r *http.Request) error {
	const q = "SELECT id, username, email, bio, avatar, created_at FROM users WHERE id = $1"
	row, err := s.dbClient.QueryRow(r.Context(), q, r.Context().Value(middleware.CtxUserKey))
	if err != nil {
		return err
	}
	// I'm using row.Scan instead of QueryRowToStruct to avoid having to deal with
	// passwords/password hashes
	var user User
	row.Scan(&user.ID, &user.Username, &user.Email, &user.Bio, &user.Avatar, &user.CreatedAt)
	return json.NewEncoder(w).Encode(user)
}

func (s *Server) apiHandlePostMe(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	type userUpdate struct {
		Bio    string
		Avatar int
	}
	var update userUpdate
	if err := json.Unmarshal(reqBody, &update); err != nil {
		return err
	}

	q := `UPDATE users SET bio = $1, avatar = $2 WHERE id = $3
	RETURNING id, username, email, bio, avatar, created_at`
	row, err := s.dbClient.QueryRow(r.Context(), q, update.Bio, update.Avatar, r.Context().Value(middleware.CtxUserKey))
	if err != nil {
		return err
	}
	var user User
	row.Scan(&user.ID, &user.Username, &user.Email, &user.Bio, &user.Avatar, &user.CreatedAt)
	return json.NewEncoder(w).Encode(user)
}

func (s *Server) apiHandlePostComments(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var c Comment
	if err := json.Unmarshal(reqBody, &c); err != nil {
		return err
	}

	const q = `
	INSERT INTO comments (thread_id, body, author_id)
	VALUES ($1, $2, $3)
	RETURNING comment_id, thread_id, author_id, body, created_at`

	comment, err := pg.QueryRowToStruct[Comment](r.Context(), s.dbClient, q, c.ThreadID, c.Body, r.Context().Value(middleware.CtxUserKey))
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(comment)
}

func (s *Server) apiHandleGetCommentsByThreadID(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid thread ID %q", r.PathValue("id"))
	}

	q := pageBuilder(`SELECT comment_id, thread_id, author_id, body, created_at FROM comments WHERE thread_id = $1`, r)
	comments, err := pg.QueryRowsToStruct[Comment](r.Context(), s.dbClient, q, id)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(comments)
}

func (s *Server) apiHandleGetCommentByID(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return err
	}

	q := `SELECT comment_id, thread_id, author_id, body, created_at FROM comments WHERE comment_id = $1`
	comment, err := pg.QueryRowToStruct[Comment](r.Context(), s.dbClient, q, id)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(comment)
}

func (s *Server) apiHandleGetCategories(w http.ResponseWriter, r *http.Request) error {
	q := `SELECT category_id, title, description FROM categories`
	categories, err := pg.QueryRowsToStruct[Category](r.Context(), s.dbClient, q)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(categories)
}

func (s *Server) apiHandlePostCategories(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var c Category
	if err := json.Unmarshal(reqBody, &c); err != nil {
		return err
	}
	const q = `
	INSERT INTO categories (title, description, author_id)
	VALUES ($1, $2)
	RETURNING category_id, title, description, author_id, created_at`

	category, err := pg.QueryRowToStruct[Category](r.Context(), s.dbClient, q, c.Title, c.Description, r.Context().Value(middleware.CtxUserKey))
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(category)
}
