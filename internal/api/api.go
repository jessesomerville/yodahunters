package api

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"

	// "github.com/jessesomerville/yodahunters/internal/envconfig"
	"github.com/jessesomerville/yodahunters/internal/log"
	"github.com/jessesomerville/yodahunters/internal/pg"
)

func HandleRequest(r *http.Request, w http.ResponseWriter) {
	// Testing code - just reflect the POST body in a 200

	// \/api\/(.+)
	pathPattern := regexp.MustCompile(`\/api\/(.+)`)
	pathMatch := pathPattern.FindStringSubmatch(r.URL.Path)

	// There shouldn't be a way to hit this handler without a match...
	if len(pathMatch) != 2 {
		log.Errorf(r.Context(), "Request with URL %q hit API Handler and failed the path regex!\nPathMatch %s", r.URL.Path, pathMatch)
		http.Error(w, "Invalid URL Path", http.StatusInternalServerError)
		return
	}

	path := pathMatch[1]

	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to read request body!")
	}

	body := string(bodyBytes)

	fmt.Fprintf(w, "Received the following body from path %s\n\n%s", html.EscapeString(path), html.EscapeString(body))
}

func HandleGetThreads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbClient, err := pg.NewClient(ctx, "postgres")
	if err != nil {
		http.Error(w, "Failed to acquire DB client", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to acquire DB client")
	}
	defer dbClient.Close(ctx)

	var threads []Thread
	rows, err := dbClient.Query(ctx, "SELECT * FROM threads")
	if err != nil {
		http.Error(w, "Failed to query DB", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to query DB")
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

	jsonData, err := json.Marshal(threads)
	if err != nil {
		http.Error(w, "Failed to marshal threads to JSON", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to marshal threads to JSON %v", threads)
	}

	w.Write(jsonData)
}

func HandleGetThreadByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error retrieving ID from Path", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Error retrieving ID from Path, ID: %d", id)
	}

	ctx := r.Context()

	dbClient, err := pg.NewClient(ctx, "postgres")
	if err != nil {
		http.Error(w, "Failed to acquire DB client", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to acquire DB client")
	}
	defer dbClient.Close(ctx)

	var thread Thread
	err = dbClient.QueryRow(ctx, "SELECT id, title, body FROM threads WHERE id = $1", id).Scan(&thread.ID, &thread.Title, &thread.Body)
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

func HandlePostThreads(w http.ResponseWriter, r *http.Request) {
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
	dbClient, err := pg.NewClient(ctx, "postgres")
	if err != nil {
		http.Error(w, "Failed to acquire DB client", http.StatusInternalServerError)
		log.Errorf(r.Context(), "Failed to acquire DB client")
	}
	defer dbClient.Close(ctx)

	var thread Thread
	err = dbClient.QueryRow(ctx, insertThreadQuery, t.Title, t.Body).Scan(&thread.ID, &thread.Title, &thread.Body)
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
