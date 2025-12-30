package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jessesomerville/yodahunters/internal/pg"
	"github.com/jessesomerville/yodahunters/internal/server/middleware"
)

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) error {
	// Handle paging
	var page middleware.Page
	page = r.Context().Value(middleware.CtxPageKey).(middleware.Page)
	offset := strconv.Itoa(page.Size * (page.Number - 1))
	size := strconv.Itoa(page.Size)

	// For each thread, we need: TODO[Category], Title, Author Name, Number of Replies, TODO[Rating], Latest Comment
	type threadView struct {
		// Category string
		Title      string `db:"title"`
		AuthorName string `db:"username"`
		ReplyCount int    `db:"reply_count"`
		// Rating int
		LatestComment string    `db:"latest_comment"`
		LatestTS      time.Time `db:"latest_ts"`
	}
	// Create a SQL query that gives us the right rows from each table
	q := `
	SELECT 
		threads.title,
		users.username,
		(SELECT COUNT(*) FROM comments WHERE comments.thread_id = threads.thread_id) AS reply_count,
		COALESCE((SELECT comments.body FROM comments WHERE comments.thread_id = threads.thread_id ORDER BY comments.created_at DESC LIMIT 1), 'No comments yet!') AS latest_comment,
		COALESCE((SELECT comments.created_at FROM comments WHERE comments.thread_id = threads.thread_id ORDER BY comments.created_at DESC LIMIT 1), threads.created_at) AS latest_ts
	FROM threads
	JOIN users ON threads.author_id = users.id
	LEFT JOIN comments ON threads.thread_id = comments.thread_id
	GROUP BY threads.thread_id, comments.created_at, users.username
	ORDER BY latest_ts DESC
	OFFSET $1 LIMIT $2`

	threadViews, err := pg.QueryRowsToStruct[threadView](r.Context(), s.dbClient, q, offset, size)
	if err != nil {
		return err
	}
	data := struct {
		ThreadViews []threadView
		HTMLTitle   string
	}{
		ThreadViews: threadViews,
		HTMLTitle:   "home",
	}

	err = s.serveHTML(r.Context(), w, "home", data)
	return err
}
