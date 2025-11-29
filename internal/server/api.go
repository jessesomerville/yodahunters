package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jessesomerville/yodahunters/internal/log"
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
