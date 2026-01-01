package server

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jessesomerville/yodahunters/internal/pg"
	"github.com/jessesomerville/yodahunters/internal/server/middleware"
)

// PageData is a struct that stores the data necessary to handle
// paging in our templates.
type PageData struct {
	PageNumber int
	PageSize   int
	Pages      []int
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) error {
	// Handle paging
	var page middleware.Page
	page = r.Context().Value(middleware.CtxPageKey).(middleware.Page)
	offset := strconv.Itoa(page.Size * (page.Number - 1))
	size := strconv.Itoa(page.Size)

	q := `SELECT COUNT(*) FROM threads`
	var threadCount int
	row, err := s.dbClient.QueryRow(r.Context(), q)
	if err != nil {
		return err
	}
	row.Scan(&threadCount)
	pages := make([]int, int(math.Ceil(float64(threadCount)/float64(page.Size))))
	for i := range pages {
		pages[i] = i + 1
	}

	// For each thread, we need: CategoryID, Title, Author Name, Number of Replies, TODO[Rating], Latest Comment
	type threadView struct {
		CategoryID int    `db:"category_id"`
		Title      string `db:"title"`
		AuthorName string `db:"username"`
		ReplyCount int    `db:"reply_count"`
		// Rating int
		LatestComment string    `db:"latest_comment"`
		LatestTS      time.Time `db:"latest_ts"`
		LatestTSFmt   string    `db:"-"`
	}
	// Create a SQL query that gives us the right rows from each table
	q = `
	SELECT 
		threads.category_id,
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
	for i := range threadViews {
		threadViews[i].LatestTSFmt = threadViews[i].LatestTS.Format("Jan 2 2006 03:04:05 PM")
	}

	data := struct {
		ThreadViews []threadView
		HTMLTitle   string
		PageData    PageData
	}{
		ThreadViews: threadViews,
		HTMLTitle:   "home",
		PageData: PageData{
			PageNumber: page.Number,
			PageSize:   page.Size,
			Pages:      pages,
		},
	}

	err = s.serveHTML(r.Context(), w, "home", data)
	return err
}

func (s *Server) handleNewThread(w http.ResponseWriter, r *http.Request) error {

	// I think it's simpler to just make entire Category structs as opposed to
	// defining a custom struct with just id and title to hold the data we need.
	q := `SELECT category_id, title, description, author_id, created_at FROM categories`
	categoryData, err := pg.QueryRowsToStruct[Category](r.Context(), s.dbClient, q)
	if err != nil {
		return err
	}

	data := struct {
		CategoryData []Category
		HTMLTitle    string
	}{
		HTMLTitle:    "new thread",
		CategoryData: categoryData,
	}

	err = s.serveHTML(r.Context(), w, "new_thread", data)
	return err
}
