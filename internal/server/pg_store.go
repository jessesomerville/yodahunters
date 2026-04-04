package server

import (
	"context"

	"github.com/jessesomerville/yodahunters/internal/pg"
	"github.com/jessesomerville/yodahunters/internal/server/middleware"
)

// Compile-time check that *pgStore implements Store.
var _ Store = (*pgStore)(nil)

type pgStore struct {
	db pg.DB
}

func newPGStore(db pg.DB) *pgStore {
	return &pgStore{db: db}
}

// --- Categories ---

func (s *pgStore) GetCategories(ctx context.Context) ([]Category, error) {
	const q = `SELECT category_id, title, description, author_id, created_at FROM categories`
	return pg.QueryRowsToStruct[Category](ctx, s.db, q)
}

func (s *pgStore) CreateCategory(ctx context.Context, title, description string, authorID int) (Category, error) {
	const q = `
	INSERT INTO categories (title, description, author_id)
	VALUES ($1, $2, $3)
	RETURNING category_id, title, description, author_id, created_at`
	return pg.QueryRowToStruct[Category](ctx, s.db, q, title, description, authorID)
}

// --- Comments ---

func (s *pgStore) GetCommentByID(ctx context.Context, id int) (Comment, error) {
	const q = `SELECT comment_id, thread_id, author_id, body, created_at FROM comments WHERE comment_id = $1`
	return pg.QueryRowToStruct[Comment](ctx, s.db, q, id)
}

func (s *pgStore) GetCommentsByThreadID(ctx context.Context, threadID int, page middleware.Page) ([]Comment, error) {
	const q = `
	SELECT comment_id, thread_id, author_id, body, created_at
	FROM comments WHERE thread_id = $1
	ORDER BY created_at DESC OFFSET $2 LIMIT $3`
	return pg.QueryRowsToStruct[Comment](ctx, s.db, q, threadID, page.Size*(page.Number-1), page.Size)
}

func (s *pgStore) GetCommentCountByThreadID(ctx context.Context, threadID int) (int, error) {
	row, err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM comments WHERE thread_id = $1`, threadID)
	if err != nil {
		return 0, err
	}
	var count int
	return count, row.Scan(&count)
}

func (s *pgStore) GetCommentViews(ctx context.Context, threadID int, page middleware.Page) ([]CommentView, error) {
	const q = `
	SELECT
		c1.author_id,
		users.avatar,
		users.username,
		c1.comment_id,
		c1.reply_id,
		c1.body,
		COALESCE((SELECT ind FROM (SELECT c1.comment_id, ROW_NUMBER() OVER (ORDER BY c1.created_at ASC) AS ind) WHERE c1.reply_id = c2.comment_id) / $3 + 1, -1) AS reply_page,
		COALESCE((SELECT body FROM comments WHERE comments.comment_id = c1.reply_id), '') AS reply_body,
		COALESCE((SELECT username FROM users WHERE id = c2.author_id), '') AS reply_author_username,
		COALESCE((SELECT id FROM users WHERE id = c2.author_id), -1) AS reply_author_id,
		c1.created_at
	FROM comments AS c1
	JOIN users ON c1.author_id = users.id
	LEFT JOIN comments AS c2 ON c1.reply_id = c2.comment_id
	WHERE c1.thread_id = $1
	ORDER BY c1.created_at ASC
	OFFSET $2 LIMIT $3`
	return pg.QueryRowsToStruct[CommentView](ctx, s.db, q, threadID, page.Size*(page.Number-1), page.Size)
}

func (s *pgStore) CreateComment(ctx context.Context, threadID, replyID, authorID int, body string) (Comment, error) {
	const q = `
	INSERT INTO comments (thread_id, body, reply_id, author_id)
	VALUES ($1, $2, $3, $4)
	RETURNING comment_id, thread_id, author_id, body, reply_id, created_at`
	return pg.QueryRowToStruct[Comment](ctx, s.db, q, threadID, body, replyID, authorID)
}

// --- Threads ---

func (s *pgStore) GetThreads(ctx context.Context, page middleware.Page) ([]Thread, error) {
	const q = `
	SELECT thread_id, author_id, category_id, title, body, created_at FROM threads
	ORDER BY created_at DESC OFFSET $1 LIMIT $2`
	return pg.QueryRowsToStruct[Thread](ctx, s.db, q, page.Size*(page.Number-1), page.Size)
}

func (s *pgStore) GetThreadsByCategoryID(ctx context.Context, categoryID int, page middleware.Page) ([]Thread, error) {
	const q = `
	SELECT thread_id, author_id, category_id, title, body, created_at FROM threads
	WHERE category_id = $1
	ORDER BY created_at DESC OFFSET $2 LIMIT $3`
	return pg.QueryRowsToStruct[Thread](ctx, s.db, q, categoryID, page.Size*(page.Number-1), page.Size)
}

func (s *pgStore) GetThreadByID(ctx context.Context, id int) (Thread, error) {
	const q = `SELECT thread_id, author_id, category_id, title, body, created_at FROM threads WHERE thread_id = $1`
	return pg.QueryRowToStruct[Thread](ctx, s.db, q, id)
}

func (s *pgStore) GetThreadCount(ctx context.Context) (int, error) {
	row, err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM threads`)
	if err != nil {
		return 0, err
	}
	var count int
	return count, row.Scan(&count)
}

func (s *pgStore) GetThreadCountByCategory(ctx context.Context, categoryID int) (int, error) {
	row, err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM threads WHERE category_id = $1`, categoryID)
	if err != nil {
		return 0, err
	}
	var count int
	return count, row.Scan(&count)
}

func (s *pgStore) GetHomeThreadViews(ctx context.Context, page middleware.Page) ([]ThreadView, error) {
	const q = `
	SELECT * FROM (
		SELECT DISTINCT ON (threads.thread_id)
			threads.category_id,
			threads.thread_id,
			threads.title,
			threads.author_id,
			threads.pinned,
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
	return pg.QueryRowsToStruct[ThreadView](ctx, s.db, q, page.Size*(page.Number-1), page.Size)
}

func (s *pgStore) GetPinnedThreadViews(ctx context.Context) ([]ThreadView, error) {
	const q = `
	SELECT DISTINCT ON (threads.thread_id)
		threads.category_id,
		threads.thread_id,
		threads.title,
		threads.author_id,
		threads.pinned,
		users.username,
		(SELECT COUNT(*) FROM comments WHERE comments.thread_id = threads.thread_id) AS reply_count,
		COALESCE((SELECT comments.body FROM comments WHERE comments.thread_id = threads.thread_id ORDER BY comments.created_at DESC LIMIT 1), 'No comments yet!') AS latest_comment,
		COALESCE((SELECT comments.comment_id FROM comments WHERE comments.thread_id = threads.thread_id ORDER BY comments.created_at DESC LIMIT 1), 0) AS latest_comment_id,
		COALESCE((SELECT comments.created_at FROM comments WHERE comments.thread_id = threads.thread_id ORDER BY comments.created_at DESC LIMIT 1), threads.created_at) AS latest_ts
	FROM threads
	JOIN users ON threads.author_id = users.id
	LEFT JOIN comments ON threads.thread_id = comments.thread_id
	WHERE threads.pinned = true
	GROUP BY threads.thread_id, comments.created_at, users.username
	ORDER BY threads.thread_id DESC`
	return pg.QueryRowsToStruct[ThreadView](ctx, s.db, q)
}

func (s *pgStore) GetCategoryThreadViews(ctx context.Context, categoryID int, page middleware.Page) ([]ThreadView, error) {
	const q = `
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
		WHERE threads.category_id = $1
		GROUP BY threads.thread_id, comments.created_at, users.username
		ORDER BY threads.thread_id DESC)
	ORDER BY latest_ts DESC
	OFFSET $2 LIMIT $3`
	return pg.QueryRowsToStruct[ThreadView](ctx, s.db, q, categoryID, page.Size*(page.Number-1), page.Size)
}

func (s *pgStore) GetThreadDetail(ctx context.Context, threadID int) (ThreadDetail, error) {
	const q = `
	SELECT
		threads.title, threads.thread_id, threads.body, threads.author_id,
		users.avatar, users.username, threads.category_id,
		categories.title AS category_title, threads.created_at
	FROM threads
	JOIN users ON threads.author_id = users.id
	JOIN categories ON threads.category_id = categories.category_id
	WHERE thread_id = $1`
	return pg.QueryRowToStruct[ThreadDetail](ctx, s.db, q, threadID)
}

func (s *pgStore) CreateThread(ctx context.Context, title, body string, categoryID, authorID int) (Thread, error) {
	const q = `
	INSERT INTO threads (title, body, category_id, author_id)
	VALUES ($1, $2, $3, $4)
	RETURNING thread_id, author_id, category_id, title, body, created_at`
	return pg.QueryRowToStruct[Thread](ctx, s.db, q, title, body, categoryID, authorID)
}

// --- Users ---

func (s *pgStore) GetUserByID(ctx context.Context, id int) (User, error) {
	const q = `SELECT id, username, bio, avatar, created_at, is_admin FROM users WHERE id = $1`
	return pg.QueryRowToStruct[User](ctx, s.db, q, id)
}

func (s *pgStore) GetUserForLogin(ctx context.Context, username string) (User, error) {
	const q = `SELECT id, pw_hash, is_admin FROM users WHERE username = $1`
	return pg.QueryRowToStruct[User](ctx, s.db, q, username)
}

func (s *pgStore) UpdateUser(ctx context.Context, bio string, avatar, userID int) (User, error) {
	const q = `
	UPDATE users SET bio = $1, avatar = $2 WHERE id = $3
	RETURNING id, username, email, bio, avatar, created_at`
	return pg.QueryRowToStruct[User](ctx, s.db, q, bio, avatar, userID)
}

// --- Registration ---

func (s *pgStore) ValidateRegKey(ctx context.Context, regKey string) (bool, error) {
	row, err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM registration_keys WHERE reg_key = $1 AND used = false)`, regKey)
	if err != nil {
		return false, err
	}
	var exists bool
	return exists, row.Scan(&exists)
}

func (s *pgStore) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	row, err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username)
	if err != nil {
		return false, err
	}
	var exists bool
	return exists, row.Scan(&exists)
}

func (s *pgStore) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	row, err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email)
	if err != nil {
		return false, err
	}
	var exists bool
	return exists, row.Scan(&exists)
}

func (s *pgStore) CreateUser(ctx context.Context, username, email string, pwHash []byte, bio string, avatar int) (User, error) {
	const q = `
	INSERT INTO users (username, email, pw_hash, bio, avatar)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, username, email, bio, avatar, created_at`
	return pg.QueryRowToStruct[User](ctx, s.db, q, username, email, pwHash, bio, avatar)
}

func (s *pgStore) UseRegKey(ctx context.Context, userID int, regKey string) error {
	return s.db.Exec(ctx, `UPDATE registration_keys SET used = true, used_by = $1 WHERE reg_key = $2`, userID, regKey)
}
