// Package middleware implements middleware HTTP handlers.
package middleware

import (
	"context"
	"net/http"
)

type CtxKey string

const CtxUserKey CtxKey = "userID"

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

func MiddlewareChain(f func(w http.ResponseWriter, r *http.Request) error, jwtSecret []byte) http.HandlerFunc {
	return AuthorizationHandler(ErrorHandler(f), jwtSecret)
}
