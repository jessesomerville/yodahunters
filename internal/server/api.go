package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jessesomerville/yodahunters/internal/log"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) handleGetThreads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := s.dbClient.Query(ctx, "SELECT id, title, body, created_at FROM threads")
	if err != nil {
		http.Error(w, "Failed to query DB", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to query DB: %v", err)
	}
	defer rows.Close()
	threads, err := pgx.CollectRows(rows, pgx.RowToStructByName[Thread])
	if err != nil {
		http.Error(w, "Failed to scan DB results!", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to scan DB results: %v", err)
	}

	jsonData, err := json.Marshal(threads)
	if err != nil {
		http.Error(w, "Failed to marshal threads to JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to marshal threads to JSON %v", threads)
	}

	w.Write(jsonData)
}

func (s *Server) handleGetThreadByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error retrieving ID from Path", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Error retrieving ID from Path, ID: %d", id)
	}

	ctx := r.Context()

	var thread Thread
	err = s.dbClient.QueryRow(ctx, "SELECT id, title, body, created_at FROM threads WHERE id = $1", id).
		Scan(&thread.ID, &thread.Title, &thread.Body, &thread.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to retrieve thread!", http.StatusInternalServerError)
		log.Errorf(ctx, "Failed to retrieve thread with ID: %d", id)
	}

	jsonData, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, "Failed to marshal thread to JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to marshal thread to JSON %v", thread)
	}

	w.Write(jsonData)
}

func (s *Server) handlePostThreads(w http.ResponseWriter, r *http.Request) {
	const insertThreadQuery = `
	INSERT INTO threads (title, body)
	VALUES ($1, $2)
	RETURNING id, title, body, created_at`

	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to read request body!")
	}

	var t Thread
	if err = json.Unmarshal(bodyBytes, &t); err != nil {
		http.Error(w, "Failed to parse request JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to parse request JSON!")
	}

	ctx := r.Context()

	var thread Thread
	err = s.dbClient.QueryRow(ctx, insertThreadQuery, t.Title, t.Body).
		Scan(&thread.ID, &thread.Title, &thread.Body, &thread.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to update threads table", http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't update threads table!")
	}

	jsonData, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, "Failed to marshal thread to JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to marshal thread to JSON %v", thread)
	}

	w.Write(jsonData)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	const insertUserQuery = `
	INSERT INTO users (username, email, pw_hash)
	VALUES ($1, $2, $3)
	RETURNING id, username, email, created_at`

	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to read request body!")
	}

	var u User
	if err = json.Unmarshal(bodyBytes, &u); err != nil {
		http.Error(w, "Failed to parse request JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to parse request JSON!")
	}

	u.GeneratePasswordHash()
	ctx := r.Context()

	var user User
	err = s.dbClient.QueryRow(ctx, insertUserQuery, u.Username, u.Email, u.PasswordHash).
		Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to update users table", http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't update users table!")
	}

	jsonData, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Failed to marshal user to JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to marshal user to JSON %v", user)
	}

	w.Write(jsonData)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to read request body!")
	}

	var login struct {
		Username string
		Password string
	}
	if err = json.Unmarshal(bodyBytes, &login); err != nil {
		http.Error(w, "Failed to parse request JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to parse request JSON!")
	}

	ctx := r.Context()

	var user User
	userQueryString := "SELECT (id, username, email, pw_hash, created_at) FROM users WHERE username = ($1)"
	err = s.dbClient.QueryRow(ctx, userQueryString, fmt.Sprintf("%c%s%c", '\'', login.Username, '\'')).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to find user with that username", http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't find user with username: %s\n%v", login.Username, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(login.Password))
	if err != nil {
		http.Error(w, "Incorrect password", http.StatusInternalServerError)
		log.Errorf(ctx, "Incorrect password")
	}

	jsonData, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Failed to marshal user to JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to marshal user to JSON %v", user)
	}

	w.Write(jsonData)
}
