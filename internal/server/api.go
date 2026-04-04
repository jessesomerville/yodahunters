package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/jessesomerville/yodahunters/internal/server/middleware"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (s *Server) apiHandleGetThreads(w http.ResponseWriter, r *http.Request) error {
	page := r.Context().Value(middleware.CtxPageKey).(middleware.Page)
	threads, err := s.store.GetThreads(r.Context(), page)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(threads)
}

func (s *Server) getHandleGetThreadsByCategoryID(w http.ResponseWriter, r *http.Request) error {
	categoryID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid category ID %q", r.PathValue("id"))
	}
	page := r.Context().Value(middleware.CtxPageKey).(middleware.Page)
	threads, err := s.store.GetThreadsByCategoryID(r.Context(), categoryID, page)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(threads)
}

func (s *Server) apiHandleGetThreadByID(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid thread ID %q", r.PathValue("id"))
	}
	thread, err := s.store.GetThreadByID(r.Context(), id)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(thread)
}

func (s *Server) apiHandlePostThreads(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var t Thread
	if err := json.Unmarshal(reqBody, &t); err != nil {
		return err
	}
	authorID := r.Context().Value(middleware.CtxUserKey).(int)
	thread, err := s.store.CreateThread(r.Context(), t.Title, t.Body, t.CategoryID, authorID)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(thread)
}

func (s *Server) apiHandleRegister(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	type reqData struct {
		RegistrationKey string `json:"reg_key"`
		Username        string `json:"username"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		Bio             string `json:"bio"`
		Avatar          int    `json:"avatar"`
	}
	var data reqData
	if err := json.Unmarshal(reqBody, &data); err != nil {
		return err
	}

	if !emailRegex.MatchString(data.Email) {
		return fmt.Errorf("invalid email address")
	}

	regKeyExists, err := s.store.ValidateRegKey(r.Context(), data.RegistrationKey)
	if err != nil {
		return err
	}
	if !regKeyExists {
		return fmt.Errorf("invalid registration key")
	}

	userExists, err := s.store.CheckUsernameExists(r.Context(), data.Username)
	if err != nil {
		return err
	}
	if userExists {
		return fmt.Errorf("user with username: %s already exists", data.Username)
	}

	emailExists, err := s.store.CheckEmailExists(r.Context(), data.Email)
	if err != nil {
		return err
	}
	if emailExists {
		return fmt.Errorf("user with email: %s already exists", data.Email)
	}

	u := User{
		Username: data.Username,
		Email:    data.Email,
		Password: data.Password,
		Bio:      data.Bio,
		Avatar:   data.Avatar,
	}
	if err = u.GeneratePasswordHash(); err != nil {
		return err
	}

	user, err := s.store.CreateUser(r.Context(), u.Username, u.Email, u.PasswordHash, u.Bio, u.Avatar)
	if err != nil {
		return err
	}
	if err := s.store.UseRegKey(r.Context(), user.ID, data.RegistrationKey); err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(user)
}

func (s *Server) apiHandleLogin(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var login struct {
		Username string
		Password string
	}
	if err := json.Unmarshal(reqBody, &login); err != nil {
		return err
	}

	user, err := s.store.GetUserForLogin(r.Context(), login.Username)
	if err != nil {
		return err
	}
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(login.Password)); err != nil {
		return err
	}

	jwt, err := middleware.GenerateJWT(user.ID, user.IsAdmin, s.jwtSecret)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    jwt.Raw,
		Expires:  time.Now().Add(12 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   !s.devmode,
		SameSite: http.SameSiteLaxMode,
	})

	return json.NewEncoder(w).Encode(struct {
		AccessToken string `json:"access_token"`
	}{jwt.Raw})
}

func (s *Server) apiHandleGetMe(w http.ResponseWriter, r *http.Request) error {
	userID := r.Context().Value(middleware.CtxUserKey).(int)
	user, err := s.store.GetUserByID(r.Context(), userID)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(user)
}

func (s *Server) apiHandlePostMe(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var update struct {
		Bio    string
		Avatar int
	}
	if err := json.Unmarshal(reqBody, &update); err != nil {
		return err
	}
	userID := r.Context().Value(middleware.CtxUserKey).(int)
	user, err := s.store.UpdateUser(r.Context(), update.Bio, update.Avatar, userID)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(user)
}

func (s *Server) apiHandlePostComments(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var c Comment
	if err := json.Unmarshal(reqBody, &c); err != nil {
		return err
	}
	authorID := r.Context().Value(middleware.CtxUserKey).(int)
	comment, err := s.store.CreateComment(r.Context(), c.ThreadID, c.ReplyID, authorID, c.Body)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(comment)
}

func (s *Server) apiHandleGetCommentsByThreadID(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid thread ID %q", r.PathValue("id"))
	}
	page := r.Context().Value(middleware.CtxPageKey).(middleware.Page)
	comments, err := s.store.GetCommentsByThreadID(r.Context(), id, page)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(comments)
}

func (s *Server) apiHandleGetCommentByID(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid comment ID %q", r.PathValue("id"))
	}
	comment, err := s.store.GetCommentByID(r.Context(), id)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(comment)
}

func (s *Server) apiHandleGetCategories(w http.ResponseWriter, r *http.Request) error {
	categories, err := s.store.GetCategories(r.Context())
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(categories)
}

func (s *Server) apiHandlePostCategories(w http.ResponseWriter, r *http.Request) error {
	reqBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	var c Category
	if err := json.Unmarshal(reqBody, &c); err != nil {
		return err
	}
	authorID := r.Context().Value(middleware.CtxUserKey).(int)
	category, err := s.store.CreateCategory(r.Context(), c.Title, c.Description, authorID)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(category)
}
