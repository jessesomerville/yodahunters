package server

import (
	"fmt"
	"math"
	"net/http"
	"os"
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

// TODO: Refactor so there's a constructor for PageData similar to
// NewHeaderData.

// HeaderData is a stuct for holding all the bits of data
// used by all our templates.
type HeaderData struct {
	UserID    int
	HTMLTitle string
}

// NewHeaderData is a constructor for the HeaderData type.
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
		ThreadID   int    `db:"thread_id"`
		Title      string `db:"title"`
		AuthorName string `db:"username"`
		AuthorID   int    `db:"author_id"`
		ReplyCount int    `db:"reply_count"`
		// Rating int
		LatestComment   string    `db:"latest_comment"`
		LatestCommentID int       `db:"latest_comment_id"`
		LatestTS        time.Time `db:"latest_ts"`
	}
	// Create a SQL query that gives us the right rows from each table
	q = `
	SELECT * FROM (
		SELECT DISTINCT ON (threads.thread_id)
			threads.category_id,
			threads.thread_id,
			threads.title,
			threads.author_id,
			users.username,
			(SELECT COUNT(*) FROM comments WHERE comments.thread_id = threads.thread_id) AS reply_count,
			COALESCE((SELECT comments.body FROM comments WHERE comments.thread_id = threads.thread_id ORDER BY comments.created_at DESC LIMIT 1), 'No comments yet!') AS latest_comment,
			COALESCE((SELECT comments.comment_id FROM comments WHERE comments.thread_id = threads.thread_id ORDER BY comments.created_at DESC LIMIT 1), 0) AS latest_comment_id,
			COALESCE((SELECT comments.created_at FROM comments WHERE comments.thread_id = threads.thread_id ORDER BY comments.created_at DESC LIMIT 1), threads.created_at) AS latest_ts
		FROM threads
		JOIN users ON threads.author_id = users.id
		LEFT JOIN comments ON threads.thread_id = comments.thread_id
		GROUP BY threads.thread_id, comments.created_at, users.username
		ORDER BY threads.thread_id DESC)
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

func (s *Server) handleThread(w http.ResponseWriter, r *http.Request) error {
	// Handle paging
	var page middleware.Page
	page = r.Context().Value(middleware.CtxPageKey).(middleware.Page)
	offset := strconv.Itoa(page.Size * (page.Number - 1))
	size := strconv.Itoa(page.Size)

	threadID := r.PathValue("id")

	// Find the number of comments in the thread for paging
	q := `SELECT COUNT(*) FROM comments WHERE thread_id = $1`
	var commentCount int
	row, err := s.dbClient.QueryRow(r.Context(), q, threadID)
	if err != nil {
		return err
	}
	row.Scan(&commentCount)
	pages := make([]int, int(math.Ceil(float64(commentCount)/float64(page.Size))))
	for i := range pages {
		pages[i] = i + 1
	}

	type threadData struct {
		Title         string    `db:"title"`
		ThreadID      int       `db:"thread_id"`
		Body          string    `db:"body"`
		AuthorID      int       `db:"author_id"`
		Avatar        int       `db:"avatar"`
		AvatarStr     string    `db:"-"`
		Username      string    `db:"username"`
		CategoryID    int       `db:"category_id"`
		CategoryTitle string    `db:"category_title"`
		CreatedAt     time.Time `db:"created_at"`
	}

	q = `
	SELECT 
		threads.title, threads.thread_id, threads.body, threads.author_id, users.avatar, users.username, threads.category_id, categories.title AS category_title, threads.created_at 
	FROM threads 
	JOIN users ON threads.author_id = users.id
	JOIN categories ON threads.category_id = categories.category_id
	WHERE thread_id = $1`
	thread, err := pg.QueryRowToStruct[threadData](r.Context(), s.dbClient, q, threadID)
	if err != nil {
		return err
	}
	thread.AvatarStr = fmt.Sprintf("%03d", thread.Avatar)

	type commentView struct {
		AuthorID            int       `db:"author_id"`
		Avatar              int       `db:"avatar"`
		AvatarStr           string    `db:"-"`
		Username            string    `db:"username"`
		CommentID           int       `db:"comment_id"`
		ReplyID             int       `db:"reply_id"`
		ReplyPage           int       `db:"reply_page"`
		ReplyBody           string    `db:"reply_body"`
		ReplyAuthorUsername string    `db:"reply_author_username"`
		ReplyAuthorID       int       `db:"reply_author_id"`
		Body                string    `db:"body"`
		CreatedAt           time.Time `db:"created_at"`
	}

	q = `
	SELECT
		c1.author_id, 
		users.avatar, 
		users.username, 
		c1.comment_id,
		c1.reply_id,
		c1.body,
		COALESCE((SELECT ind FROM (SELECT c1.comment_id, ROW_NUMBER() OVER (ORDER BY c1.created_at ASC) AS ind) WHERE c1.reply_id = c2.comment_id) / $3 + 1, -1) AS reply_page,
		COALESCE((SELECT body FROM comments WHERE comments.comment_id = c1.reply_id ), '') AS reply_body,
		COALESCE((SELECT username FROM users WHERE id = c2.author_id ), '') AS reply_author_username,
		COALESCE((SELECT id FROM users WHERE id = c2.author_id ), -1) AS reply_author_id,
		c1.created_at
	FROM comments AS c1
	JOIN users ON c1.author_id = users.id
	LEFT JOIN comments AS c2 ON c1.reply_id = c2.comment_id
	WHERE c1.thread_id = $1
	ORDER BY c1.created_at ASC
	OFFSET $2 LIMIT $3`

	commentViews, err := pg.QueryRowsToStruct[commentView](r.Context(), s.dbClient, q, threadID, offset, size)
	if err != nil {
		return err
	}
	for i := range commentViews {
		commentViews[i].AvatarStr = fmt.Sprintf("%03d", commentViews[i].Avatar)
	}

	data := struct {
		ThreadData   threadData
		CommentViews []commentView
		HeaderData   HeaderData
		PageData     PageData
	}{
		ThreadData:   thread,
		CommentViews: commentViews,
		HeaderData:   NewHeaderData(thread.Title, r),
		PageData: PageData{
			PageNumber: page.Number,
			PageSize:   page.Size,
			Pages:      pages,
		},
	}

	err = s.serveHTML(r.Context(), w, "thread", data)
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
	q := `SELECT id, username, bio, avatar, created_at, is_admin FROM users WHERE id = $1`
	var user User
	var isAdmin bool
	row, err := s.dbClient.QueryRow(r.Context(), q, r.PathValue("id"))
	if err != nil {
		return err
	}
	row.Scan(&user.ID, &user.Username, &user.Bio, &user.Avatar, &user.CreatedAt, &isAdmin)
	if user.Username == "" {
		return fmt.Errorf("user with id %q not found", r.PathValue("id"))
	}

	// Passing the avatar id as a string with two leading zeros for use in the template.
	// This will likely need to be changed when we update to better profile pics.
	data := struct {
		Username       string
		Bio            string
		Avatar         string
		IsAdmin        bool
		CreatedAt      time.Time
		HeaderData     HeaderData
		ShowEditButton bool
	}{
		HeaderData:     NewHeaderData(user.Username, r),
		Username:       user.Username,
		Bio:            user.Bio,
		Avatar:         fmt.Sprintf("%03d", user.Avatar),
		IsAdmin:        isAdmin,
		CreatedAt:      user.CreatedAt,
		ShowEditButton: r.Context().Value(middleware.CtxUserKey).(int) == user.ID,
	}
	err = s.serveHTML(r.Context(), w, "users", data)
	return err
}

func (s *Server) handleUsersEdit(w http.ResponseWriter, r *http.Request) error {
	// Query user info for the logged in user.
	q := `SELECT id, username, bio, avatar, created_at, is_admin FROM users WHERE id = $1`
	var user User
	var isAdmin bool
	row, err := s.dbClient.QueryRow(r.Context(), q, r.Context().Value(middleware.CtxUserKey).(int))
	if err != nil {
		return err
	}
	row.Scan(&user.ID, &user.Username, &user.Bio, &user.Avatar, &user.CreatedAt, &isAdmin)
	if user.Username == "" {
		return fmt.Errorf("user with id %q not found", r.PathValue("id"))
	}

	// This is a little bit hacky, but it makes managing profile pics
	// very easy for now.
	entries, err := os.ReadDir("static/img/pfps")
	if err != nil {
		return err
	}
	avatarNumbers := make([]string, len(entries))
	for i := range len(entries) {
		avatarNumbers[i] = fmt.Sprintf("%03d", i)
	}
	data := struct {
		Username      string
		Bio           string
		Avatar        string
		IsAdmin       bool
		CreatedAt     time.Time
		HeaderData    HeaderData
		AvatarNumbers []string
	}{
		HeaderData:    NewHeaderData(user.Username, r),
		Username:      user.Username,
		Bio:           user.Bio,
		Avatar:        fmt.Sprintf("%03d", user.Avatar),
		IsAdmin:       isAdmin,
		CreatedAt:     user.CreatedAt,
		AvatarNumbers: avatarNumbers,
	}
	err = s.serveHTML(r.Context(), w, "edit_profile", data)
	return err
}
