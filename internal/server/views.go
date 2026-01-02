package server

import (
	"fmt"
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

// HeaderData is a stuct for holding all the bits of data
// used by all our templates.
type HeaderData struct {
	UserID    int
	HTMLTitle string
}

func NewHeaderData(title string, r *http.Request) HeaderData {
	return HeaderData{
		HTMLTitle: title,
		UserID:    r.Context().Value(middleware.CtxUserKey).(int),
	}
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

	data := struct {
		ThreadViews []threadView
		HeaderData  HeaderData
		PageData    PageData
	}{
		ThreadViews: threadViews,
		HeaderData:  NewHeaderData("home", r),
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
		HeaderData   HeaderData
	}{
		HeaderData:   NewHeaderData("new thread", r),
		CategoryData: categoryData,
	}

	err = s.serveHTML(r.Context(), w, "new_thread", data)
	return err
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) error {
	q := `SELECT username, bio, avatar, created_at, is_admin FROM users WHERE id = $1`
	var user User
	var isAdmin bool
	row, err := s.dbClient.QueryRow(r.Context(), q, r.PathValue("id"))
	if err != nil {
		return err
	}
	row.Scan(&user.Username, &user.Bio, &user.Avatar, &user.CreatedAt, &isAdmin)
	if user.Username == "" {
		return fmt.Errorf("user with id %q not found", r.PathValue("id"))
	}

	// Passing the avatar id as a string with two leading zeros for use in the template.
	// This will likely need to be changed when we update to better profile pics.
	data := struct {
		Username   string
		Bio        string
		Avatar     string
		IsAdmin    bool
		CreatedAt  time.Time
		HeaderData HeaderData
	}{
		HeaderData: NewHeaderData(user.Username, r),
		Username:   user.Username,
		Bio:        user.Bio,
		Avatar:     fmt.Sprintf("%03d", user.Avatar),
		IsAdmin:    isAdmin,
		CreatedAt:  user.CreatedAt,
	}
	err = s.serveHTML(r.Context(), w, "users", data)
	return err
}
