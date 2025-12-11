// Package middleware implements middleware HTTP handlers.
package middleware

import (
	"context"
	"net/http"
)

type ctxKey string

// CtxUserKey is used to set and retrieve the user ID from a request context.
const CtxUserKey ctxKey = "userID"

// AuthorizationHandler verifies that a request has a valid access token in the
// cookie, retrieves the user_id set in the access token, and adds the user_id
// to the request context.
func AuthorizationHandler(next http.HandlerFunc, secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := Authorize(r, secret)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		ctx := context.WithValue(r.Context(), CtxUserKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Chain links all the middleware handlers together to make it convenient to call them all
// in a row.
func Chain(f func(w http.ResponseWriter, r *http.Request) error, jwtSecret []byte) http.HandlerFunc {
	return AuthorizationHandler(ErrorHandler(f), jwtSecret)
}
