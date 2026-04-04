package server

import (
	"context"

	"github.com/jessesomerville/yodahunters/internal/server/middleware"
)

// Store is the interface for all database operations performed by the server.
type Store interface {
	// Categories
	GetCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, title, description string, authorID int) (Category, error)

	// Comments
	GetCommentByID(ctx context.Context, id int) (Comment, error)
	GetCommentsByThreadID(ctx context.Context, threadID int, page middleware.Page) ([]Comment, error)
	GetCommentCountByThreadID(ctx context.Context, threadID int) (int, error)
	GetCommentViews(ctx context.Context, threadID int, page middleware.Page) ([]CommentView, error)
	CreateComment(ctx context.Context, threadID, replyID, authorID int, body string) (Comment, error)

	// Threads
	GetThreads(ctx context.Context, page middleware.Page) ([]Thread, error)
	GetThreadsByCategoryID(ctx context.Context, categoryID int, page middleware.Page) ([]Thread, error)
	GetThreadByID(ctx context.Context, id int) (Thread, error)
	GetThreadCount(ctx context.Context) (int, error)
	GetThreadCountByCategory(ctx context.Context, categoryID int) (int, error)
	GetHomeThreadViews(ctx context.Context, page middleware.Page) ([]ThreadView, error)
	GetPinnedThreadViews(ctx context.Context) ([]ThreadView, error)
	GetCategoryThreadViews(ctx context.Context, categoryID int, page middleware.Page) ([]ThreadView, error)
	GetThreadDetail(ctx context.Context, threadID int) (ThreadDetail, error)
	CreateThread(ctx context.Context, title, body string, categoryID, authorID int) (Thread, error)

	// Users
	GetUserByID(ctx context.Context, id int) (User, error)
	GetUserForLogin(ctx context.Context, username string) (User, error)
	UpdateUser(ctx context.Context, bio string, avatar, userID int) (User, error)

	// Registration
	ValidateRegKey(ctx context.Context, regKey string) (bool, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, username, email string, pwHash []byte, bio string, avatar int) (User, error)
	UseRegKey(ctx context.Context, userID int, regKey string) error
}
