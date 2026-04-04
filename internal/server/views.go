package server

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jessesomerville/yodahunters/internal/server/middleware"
	"github.com/jessesomerville/yodahunters/static"
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
	UserID     int
	HTMLTitle  string
	Categories []Category
}

// newHeaderData is a constructor for the HeaderData type.
func (s *Server) newHeaderData(title string, r *http.Request) (HeaderData, error) {
	categories, err := s.store.GetCategories(r.Context())
	if err != nil {
		return HeaderData{}, err
	}
	return HeaderData{
		HTMLTitle:  title,
		UserID:     r.Context().Value(middleware.CtxUserKey).(int),
		Categories: categories,
	}, nil
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) error {
	return s.serveHTML(r.Context(), w, "login", nil)
}

// This route is just a box to enter a regkey. It redirects to the route
// below when you submit.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) error {
	return s.serveHTML(r.Context(), w, "register", nil)
}

// This route is where you actually fill everything out to register.
func (s *Server) handleRegisterKey(w http.ResponseWriter, r *http.Request) error {
	regKey := r.PathValue("regkey")

	regKeyExists, err := s.store.ValidateRegKey(r.Context(), regKey)
	if err != nil {
		return err
	}
	if !regKeyExists {
		return fmt.Errorf("invalid registration key")
	}

	entries, err := static.FS.ReadDir("img/pfps")
	if err != nil {
		return err
	}
	avatarNumbers := make([]string, len(entries))
	for i := range len(entries) {
		avatarNumbers[i] = fmt.Sprintf("%03d", i)
	}

	data := struct {
		HeaderData    HeaderData
		AvatarNumbers []string
		RegKey        string
	}{
		HeaderData:    HeaderData{HTMLTitle: "Register"},
		AvatarNumbers: avatarNumbers,
		RegKey:        regKey,
	}
	return s.serveHTML(r.Context(), w, "register_key", data)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) error {
	page := r.Context().Value(middleware.CtxPageKey).(middleware.Page)

	threadCount, err := s.store.GetThreadCount(r.Context())
	if err != nil {
		return err
	}
	pages := make([]int, int(math.Ceil(float64(threadCount)/float64(page.Size))))
	for i := range pages {
		pages[i] = i + 1
	}

	threadViews, err := s.store.GetHomeThreadViews(r.Context(), page)
	if err != nil {
		return err
	}

	pinnedThreadViews, err := s.store.GetPinnedThreadViews(r.Context())
	if err != nil {
		return err
	}

	headerData, err := s.newHeaderData("home", r)
	if err != nil {
		return err
	}

	data := struct {
		ThreadViews   []ThreadView
		PinnedThreads []ThreadView
		HeaderData    HeaderData
		PageData      PageData
	}{
		ThreadViews:   threadViews,
		PinnedThreads: pinnedThreadViews,
		HeaderData:    headerData,
		PageData: PageData{
			PageNumber: page.Number,
			PageSize:   page.Size,
			Pages:      pages,
		},
	}
	return s.serveHTML(r.Context(), w, "home", data)
}

func (s *Server) handleCategory(w http.ResponseWriter, r *http.Request) error {
	page := r.Context().Value(middleware.CtxPageKey).(middleware.Page)

	catID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return err
	}

	threadCount, err := s.store.GetThreadCountByCategory(r.Context(), catID)
	if err != nil {
		return err
	}
	pages := make([]int, int(math.Ceil(float64(threadCount)/float64(page.Size))))
	for i := range pages {
		pages[i] = i + 1
	}

	threadViews, err := s.store.GetCategoryThreadViews(r.Context(), catID, page)
	if err != nil {
		return err
	}

	headerData, err := s.newHeaderData("home", r)
	if err != nil {
		return err
	}

	data := struct {
		ThreadViews []ThreadView
		HeaderData  HeaderData
		PageData    PageData
		CategoryID  int
	}{
		ThreadViews: threadViews,
		HeaderData:  headerData,
		CategoryID:  catID,
		PageData: PageData{
			PageNumber: page.Number,
			PageSize:   page.Size,
			Pages:      pages,
		},
	}
	return s.serveHTML(r.Context(), w, "category", data)
}

func (s *Server) handleThread(w http.ResponseWriter, r *http.Request) error {
	page := r.Context().Value(middleware.CtxPageKey).(middleware.Page)

	threadID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid thread ID %q", r.PathValue("id"))
	}

	commentCount, err := s.store.GetCommentCountByThreadID(r.Context(), threadID)
	if err != nil {
		return err
	}
	pages := make([]int, int(math.Ceil(float64(commentCount)/float64(page.Size))))
	for i := range pages {
		pages[i] = i + 1
	}

	thread, err := s.store.GetThreadDetail(r.Context(), threadID)
	if err != nil {
		return err
	}
	thread.AvatarStr = fmt.Sprintf("%03d", thread.Avatar)

	commentViews, err := s.store.GetCommentViews(r.Context(), threadID, page)
	if err != nil {
		return err
	}
	for i := range commentViews {
		commentViews[i].AvatarStr = fmt.Sprintf("%03d", commentViews[i].Avatar)
	}

	headerData, err := s.newHeaderData(thread.Title, r)
	if err != nil {
		return err
	}

	data := struct {
		ThreadData   ThreadDetail
		CommentViews []CommentView
		HeaderData   HeaderData
		PageData     PageData
	}{
		ThreadData:   thread,
		CommentViews: commentViews,
		HeaderData:   headerData,
		PageData: PageData{
			PageNumber: page.Number,
			PageSize:   page.Size,
			Pages:      pages,
		},
	}
	return s.serveHTML(r.Context(), w, "thread", data)
}

func (s *Server) handleNewThread(w http.ResponseWriter, r *http.Request) error {
	categoryData, err := s.store.GetCategories(r.Context())
	if err != nil {
		return err
	}

	headerData, err := s.newHeaderData("new thread", r)
	if err != nil {
		return err
	}

	data := struct {
		CategoryData []Category
		HeaderData   HeaderData
	}{
		HeaderData:   headerData,
		CategoryData: categoryData,
	}
	return s.serveHTML(r.Context(), w, "new_thread", data)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) error {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid user ID %q", r.PathValue("id"))
	}

	user, err := s.store.GetUserByID(r.Context(), userID)
	if err != nil {
		return err
	}
	if user.Username == "" {
		return fmt.Errorf("user with id %d not found", userID)
	}

	headerData, err := s.newHeaderData(user.Username, r)
	if err != nil {
		return err
	}

	data := struct {
		Username       string
		Bio            string
		Avatar         string
		IsAdmin        bool
		CreatedAt      time.Time
		HeaderData     HeaderData
		ShowEditButton bool
	}{
		HeaderData:     headerData,
		Username:       user.Username,
		Bio:            user.Bio,
		Avatar:         fmt.Sprintf("%03d", user.Avatar),
		IsAdmin:        user.IsAdmin,
		CreatedAt:      user.CreatedAt,
		ShowEditButton: r.Context().Value(middleware.CtxUserKey).(int) == user.ID,
	}
	return s.serveHTML(r.Context(), w, "users", data)
}

func (s *Server) handleUsersEdit(w http.ResponseWriter, r *http.Request) error {
	userID := r.Context().Value(middleware.CtxUserKey).(int)

	user, err := s.store.GetUserByID(r.Context(), userID)
	if err != nil {
		return err
	}
	if user.Username == "" {
		return fmt.Errorf("user with id %d not found", userID)
	}

	entries, err := static.FS.ReadDir("img/pfps")
	if err != nil {
		return err
	}
	avatarNumbers := make([]string, len(entries))
	for i := range len(entries) {
		avatarNumbers[i] = fmt.Sprintf("%03d", i)
	}

	headerData, err := s.newHeaderData(user.Username, r)
	if err != nil {
		return err
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
		HeaderData:    headerData,
		Username:      user.Username,
		Bio:           user.Bio,
		Avatar:        fmt.Sprintf("%03d", user.Avatar),
		IsAdmin:       user.IsAdmin,
		CreatedAt:     user.CreatedAt,
		AvatarNumbers: avatarNumbers,
	}
	return s.serveHTML(r.Context(), w, "edit_profile", data)
}
