package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jessesomerville/yodahunters/internal/pg"
	"github.com/jessesomerville/yodahunters/internal/server/middleware"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) apiHandleGetThreads(w http.ResponseWriter, r *http.Request) error {
	q := `SELECT thread_id, author_id, title, body, created_at FROM threads`
	threads, err := pg.QueryRowsToStruct[Thread](r.Context(), s.dbClient, q)
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

	q := `SELECT thread_id, author_id, title, body, created_at FROM threads WHERE id = $1`
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
	INSERT INTO threads (title, body, author_id)
	VALUES ($1, $2, $3)
	RETURNING thread_id, author_id, title, body, created_at`

	thread, err := pg.QueryRowToStruct[Thread](r.Context(), s.dbClient, q, t.Title, t.Body, r.Context().Value(middleware.CtxUserKey))
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
	INSERT INTO users (username, email, pw_hash)
	VALUES ($1, $2, $3)
	RETURNING id, username, email, created_at`
	row, err = s.dbClient.QueryRow(r.Context(), insertUser, u.Username, u.Email, u.PasswordHash)
	if err != nil {
		return err
	}
	var user User
	row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
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

	const q = "SELECT id, pw_hash FROM users WHERE username = $1"
	row, err := s.dbClient.QueryRow(r.Context(), q, login.Username)
	if err != nil {
		return err
	}
	var id int
	var passwordHash []byte
	if err = row.Scan(&id, &passwordHash); err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(login.Password))
	if err != nil {
		return err
	}

	jwt, err := middleware.GenerateJWT(id, s.jwtSecret)
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

func (s *Server) apiHandleMe(w http.ResponseWriter, r *http.Request) error {
	const q = "SELECT id, username, email, created_at FROM users WHERE id = $1"
	row, err := s.dbClient.QueryRow(r.Context(), q, r.Context().Value(middleware.CtxUserKey))
	if err != nil {
		return err
	}
	var user User
	row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	return json.NewEncoder(w).Encode(user)
}
