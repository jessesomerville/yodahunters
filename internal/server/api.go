package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/jessesomerville/yodahunters/internal/pg"
)

func (s *Server) handleGetThreads(w http.ResponseWriter, r *http.Request) error {
	q := `SELECT id, title, body, created_at FROM threads`
	threads, err := pg.QueryRowsToStruct[Thread](r.Context(), s.dbClient, q)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(threads)
}

func (s *Server) handleGetThreadByID(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid thread ID %q", r.PathValue("id"))
	}

	q := `SELECT id, title, body, created_at FROM threads WHERE id = $1`
	thread, err := pg.QueryRowToStruct[Thread](r.Context(), s.dbClient, q, id)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(thread)
}

func (s *Server) handlePostThreads(w http.ResponseWriter, r *http.Request) error {
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
	INSERT INTO threads (title, body)
	VALUES ($1, $2)
	RETURNING id, title, body, created_at`

	thread, err := pg.QueryRowToStruct[Thread](r.Context(), s.dbClient, q, t.Title, t.Body)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(thread)
}
