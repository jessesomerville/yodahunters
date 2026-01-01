// Package middleware implements middleware HTTP handlers.
package middleware

import (
	"context"
	"net/http"

	"github.com/jessesomerville/yodahunters/internal/log"
)

type ctxKey string

// CtxSecretKey is used to set and retrieve the JWT signing secret.
const CtxSecretKey ctxKey = "jwt_secret"

// CtxUserKey is used to set and retrieve the user ID from a request context.
const CtxUserKey ctxKey = "userID"

// CtxPageKey is used to set and retrieve page data.
const CtxPageKey ctxKey = "page"

// CtxAdminKey is used to set and retrieve the admin flag.
const CtxAdminKey ctxKey = "isAdmin"

// AuthorizationHandler verifies that a request has a valid access token in the
// cookie, retrieves the user_id set in the access token, and adds the user_id
// to the request context.
func AuthorizationHandler(next http.HandlerFunc, secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := Authorize(r, secret)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			log.Errorf(r.Context(), "Authorization Failed!")
			return
		}
		isAdmin, err := IsAdmin(r, secret)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			log.Errorf(r.Context(), "Admin authorization Failed!")
			return
		}
		ctx := context.WithValue(r.Context(), CtxUserKey, userID)
		ctx = context.WithValue(ctx, CtxAdminKey, isAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// PageHandler checks the URL for query parameters related to paging
// and makes them available in the request context.
func PageHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := GetPageData(r)
		if err != nil {
			log.Errorf(r.Context(), "Pagination Failed!")
		}
		ctx := context.WithValue(r.Context(), CtxPageKey, page)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// AdminHandler adds the is_admin flag stored in the JWT
// to the request context.
func AdminHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isAdmin := r.Context().Value(CtxAdminKey)
		if isAdmin != true {
			log.Errorf(r.Context(), "Admin authorization Failed!")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// Chain links all the middleware handlers together to make it convenient to call them all
// in a row.
func Chain(f func(w http.ResponseWriter, r *http.Request) error, jwtSecret []byte) http.HandlerFunc {
	return AuthorizationHandler(PageHandler(ErrorHandler(f)), jwtSecret)
}

func AdminChain(f func(w http.ResponseWriter, r *http.Request) error, jwtSecret []byte) http.HandlerFunc {
	return AuthorizationHandler(PageHandler(AdminHandler(ErrorHandler(f))), jwtSecret)
}
