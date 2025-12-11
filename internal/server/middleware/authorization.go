package middleware

import (
	"errors"
	"net/http"
)

// Authorize takes a request, verifies that it contains a valid
// JWT, then returns the user id for that the user in the JWT.
func Authorize(r *http.Request, secret []byte) (int, error) {
	accessToken, err := r.Cookie("access_token")
	if err != nil {
		return -1, err
	}
	jwt, err := ParseJWT(accessToken.Value)
	if err != nil {
		return -1, err
	}
	valid, err := jwt.IsValid(secret)
	if err != nil {
		return -1, err
	}
	if !valid {
		return -1, errors.New("invalid JWT")
	}

	return jwt.Payload.UserID, nil
}
