package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/jessesomerville/yodahunters/internal/server/middleware"
)

// fakeStore implements Store for use in handler tests.
// Set the *Fn fields to control what individual methods return;
// unset methods return zero values and a nil error.
type fakeStore struct {
	// Categories
	GetCategoriesFn    func(ctx context.Context) ([]Category, error)
	CreateCategoryFn   func(ctx context.Context, title, description string, authorID int) (Category, error)
	// Comments
	GetCommentByIDFn          func(ctx context.Context, id int) (Comment, error)
	GetCommentsByThreadIDFn   func(ctx context.Context, threadID int, page middleware.Page) ([]Comment, error)
	GetCommentCountByThreadIDFn func(ctx context.Context, threadID int) (int, error)
	GetCommentViewsFn         func(ctx context.Context, threadID int, page middleware.Page) ([]CommentView, error)
	CreateCommentFn           func(ctx context.Context, threadID, replyID, authorID int, body string) (Comment, error)
	// Threads
	GetThreadsFn             func(ctx context.Context, page middleware.Page) ([]Thread, error)
	GetThreadsByCategoryIDFn func(ctx context.Context, categoryID int, page middleware.Page) ([]Thread, error)
	GetThreadByIDFn          func(ctx context.Context, id int) (Thread, error)
	GetThreadCountFn         func(ctx context.Context) (int, error)
	GetThreadCountByCategoryFn func(ctx context.Context, categoryID int) (int, error)
	GetHomeThreadViewsFn     func(ctx context.Context, page middleware.Page) ([]ThreadView, error)
	GetPinnedThreadViewsFn   func(ctx context.Context) ([]ThreadView, error)
	GetCategoryThreadViewsFn func(ctx context.Context, categoryID int, page middleware.Page) ([]ThreadView, error)
	GetThreadDetailFn        func(ctx context.Context, threadID int) (ThreadDetail, error)
	CreateThreadFn           func(ctx context.Context, title, body string, categoryID, authorID int) (Thread, error)
	// Users
	GetUserByIDFn     func(ctx context.Context, id int) (User, error)
	GetUserForLoginFn func(ctx context.Context, username string) (User, error)
	UpdateUserFn      func(ctx context.Context, bio string, avatar, userID int) (User, error)
	// Registration
	ValidateRegKeyFn     func(ctx context.Context, regKey string) (bool, error)
	CheckUsernameExistsFn func(ctx context.Context, username string) (bool, error)
	CheckEmailExistsFn   func(ctx context.Context, email string) (bool, error)
	CreateUserFn         func(ctx context.Context, username, email string, pwHash []byte, bio string, avatar int) (User, error)
	UseRegKeyFn          func(ctx context.Context, userID int, regKey string) error
}

func (f *fakeStore) GetCategories(ctx context.Context) ([]Category, error) {
	if f.GetCategoriesFn != nil {
		return f.GetCategoriesFn(ctx)
	}
	return nil, nil
}
func (f *fakeStore) CreateCategory(ctx context.Context, title, description string, authorID int) (Category, error) {
	if f.CreateCategoryFn != nil {
		return f.CreateCategoryFn(ctx, title, description, authorID)
	}
	return Category{}, nil
}
func (f *fakeStore) GetCommentByID(ctx context.Context, id int) (Comment, error) {
	if f.GetCommentByIDFn != nil {
		return f.GetCommentByIDFn(ctx, id)
	}
	return Comment{}, nil
}
func (f *fakeStore) GetCommentsByThreadID(ctx context.Context, threadID int, page middleware.Page) ([]Comment, error) {
	if f.GetCommentsByThreadIDFn != nil {
		return f.GetCommentsByThreadIDFn(ctx, threadID, page)
	}
	return nil, nil
}
func (f *fakeStore) GetCommentCountByThreadID(ctx context.Context, threadID int) (int, error) {
	if f.GetCommentCountByThreadIDFn != nil {
		return f.GetCommentCountByThreadIDFn(ctx, threadID)
	}
	return 0, nil
}
func (f *fakeStore) GetCommentViews(ctx context.Context, threadID int, page middleware.Page) ([]CommentView, error) {
	if f.GetCommentViewsFn != nil {
		return f.GetCommentViewsFn(ctx, threadID, page)
	}
	return nil, nil
}
func (f *fakeStore) CreateComment(ctx context.Context, threadID, replyID, authorID int, body string) (Comment, error) {
	if f.CreateCommentFn != nil {
		return f.CreateCommentFn(ctx, threadID, replyID, authorID, body)
	}
	return Comment{}, nil
}
func (f *fakeStore) GetThreads(ctx context.Context, page middleware.Page) ([]Thread, error) {
	if f.GetThreadsFn != nil {
		return f.GetThreadsFn(ctx, page)
	}
	return nil, nil
}
func (f *fakeStore) GetThreadsByCategoryID(ctx context.Context, categoryID int, page middleware.Page) ([]Thread, error) {
	if f.GetThreadsByCategoryIDFn != nil {
		return f.GetThreadsByCategoryIDFn(ctx, categoryID, page)
	}
	return nil, nil
}
func (f *fakeStore) GetThreadByID(ctx context.Context, id int) (Thread, error) {
	if f.GetThreadByIDFn != nil {
		return f.GetThreadByIDFn(ctx, id)
	}
	return Thread{}, nil
}
func (f *fakeStore) GetThreadCount(ctx context.Context) (int, error) {
	if f.GetThreadCountFn != nil {
		return f.GetThreadCountFn(ctx)
	}
	return 0, nil
}
func (f *fakeStore) GetThreadCountByCategory(ctx context.Context, categoryID int) (int, error) {
	if f.GetThreadCountByCategoryFn != nil {
		return f.GetThreadCountByCategoryFn(ctx, categoryID)
	}
	return 0, nil
}
func (f *fakeStore) GetHomeThreadViews(ctx context.Context, page middleware.Page) ([]ThreadView, error) {
	if f.GetHomeThreadViewsFn != nil {
		return f.GetHomeThreadViewsFn(ctx, page)
	}
	return nil, nil
}
func (f *fakeStore) GetPinnedThreadViews(ctx context.Context) ([]ThreadView, error) {
	if f.GetPinnedThreadViewsFn != nil {
		return f.GetPinnedThreadViewsFn(ctx)
	}
	return nil, nil
}
func (f *fakeStore) GetCategoryThreadViews(ctx context.Context, categoryID int, page middleware.Page) ([]ThreadView, error) {
	if f.GetCategoryThreadViewsFn != nil {
		return f.GetCategoryThreadViewsFn(ctx, categoryID, page)
	}
	return nil, nil
}
func (f *fakeStore) GetThreadDetail(ctx context.Context, threadID int) (ThreadDetail, error) {
	if f.GetThreadDetailFn != nil {
		return f.GetThreadDetailFn(ctx, threadID)
	}
	return ThreadDetail{}, nil
}
func (f *fakeStore) CreateThread(ctx context.Context, title, body string, categoryID, authorID int) (Thread, error) {
	if f.CreateThreadFn != nil {
		return f.CreateThreadFn(ctx, title, body, categoryID, authorID)
	}
	return Thread{}, nil
}
func (f *fakeStore) GetUserByID(ctx context.Context, id int) (User, error) {
	if f.GetUserByIDFn != nil {
		return f.GetUserByIDFn(ctx, id)
	}
	return User{}, nil
}
func (f *fakeStore) GetUserForLogin(ctx context.Context, username string) (User, error) {
	if f.GetUserForLoginFn != nil {
		return f.GetUserForLoginFn(ctx, username)
	}
	return User{}, nil
}
func (f *fakeStore) UpdateUser(ctx context.Context, bio string, avatar, userID int) (User, error) {
	if f.UpdateUserFn != nil {
		return f.UpdateUserFn(ctx, bio, avatar, userID)
	}
	return User{}, nil
}
func (f *fakeStore) ValidateRegKey(ctx context.Context, regKey string) (bool, error) {
	if f.ValidateRegKeyFn != nil {
		return f.ValidateRegKeyFn(ctx, regKey)
	}
	return false, nil
}
func (f *fakeStore) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	if f.CheckUsernameExistsFn != nil {
		return f.CheckUsernameExistsFn(ctx, username)
	}
	return false, nil
}
func (f *fakeStore) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	if f.CheckEmailExistsFn != nil {
		return f.CheckEmailExistsFn(ctx, email)
	}
	return false, nil
}
func (f *fakeStore) CreateUser(ctx context.Context, username, email string, pwHash []byte, bio string, avatar int) (User, error) {
	if f.CreateUserFn != nil {
		return f.CreateUserFn(ctx, username, email, pwHash, bio, avatar)
	}
	return User{}, nil
}
func (f *fakeStore) UseRegKey(ctx context.Context, userID int, regKey string) error {
	if f.UseRegKeyFn != nil {
		return f.UseRegKeyFn(ctx, userID, regKey)
	}
	return nil
}

// newRequest builds a test request with the given context values pre-loaded.
// userID of 0 means no user is set in context (for unauthenticated handlers).
func newRequest(method, target string, body io.Reader, userID int) *http.Request {
	r := httptest.NewRequest(method, target, body)
	if userID != 0 {
		ctx := context.WithValue(r.Context(), middleware.CtxUserKey, userID)
		ctx = context.WithValue(ctx, middleware.CtxAdminKey, false)
		ctx = context.WithValue(ctx, middleware.CtxPageKey, middleware.Page{Size: 20, Number: 1})
		r = r.WithContext(ctx)
	}
	return r
}
