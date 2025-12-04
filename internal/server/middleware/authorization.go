package middleware

import (
	"net/http"

	"github.com/jessesomerville/yodahunters/internal/log"
)

func isAuthenticated(r *http.Request, secret []byte) bool {
	accessToken, err := r.Cookie("access_token")
	if err != nil {
		log.Errorf(r.Context(), "failed to read access_token cookie")
		return false
	}
	jwt, err := ParseJWT(accessToken.Value)
	if err != nil {
		log.Errorf(r.Context(), "failed to parse jwt in access_token cookie")
		return false
	}
	valid, err := jwt.IsValid(secret)
	if err != nil {
		log.Errorf(r.Context(), "unable to validate JWT")
		return false
	}
	return valid
}

func AuthorizationHandler(next http.HandlerFunc, secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r, secret) {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
	}
}
