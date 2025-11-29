package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/jessesomerville/yodahunters/internal/log"
)

func (s *Server) HandleGetThreads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var threads []Thread
	rows, err := s.dbClient.Query(ctx, "SELECT * FROM threads")
	if err != nil {
		http.Error(w, "Failed to query DB", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to query DB!")
	}
	for rows.Next() {
		var thread Thread
		err = rows.Scan(&thread.ID, &thread.Title, &thread.Body)
		if err != nil {
			rowData, _ := rows.Values()
			http.Error(w, "Failed to read query rows", http.StatusInternalServerError)
			log.Errorf(r.Context(), "Failed to read rows %s", rowData)
		}
		threads = append(threads, thread)
	}
	rows.Close()

	jsonData, err := json.Marshal(threads)
	if err != nil {
		http.Error(w, "Failed to marshal threads to JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to marshal threads to JSON %v", threads)
	}

	w.Write(jsonData)
}

func (s *Server) HandleGetThreadByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error retrieving ID from Path", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Error retrieving ID from Path, ID: %d", id)
	}

	ctx := r.Context()

	var thread Thread
	err = s.dbClient.QueryRow(ctx, "SELECT id, title, body FROM threads WHERE id = $1", id).Scan(&thread.ID, &thread.Title, &thread.Body)
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

func (s *Server) HandlePostThreads(w http.ResponseWriter, r *http.Request) {
	const insertThreadQuery = `
	INSERT INTO threads (title, body)
	VALUES ($1, $2)
	RETURNING id, title, body`

	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to read request body!")
	}

	var t Thread
	err = json.Unmarshal(bodyBytes, &t)

	if err != nil {
		http.Error(w, "Failed to parse request JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to parse request JSON!")
	}

	ctx := r.Context()

	var thread Thread
	err = s.dbClient.QueryRow(ctx, insertThreadQuery, t.Title, t.Body).Scan(&thread.ID, &thread.Title, &thread.Body)
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
