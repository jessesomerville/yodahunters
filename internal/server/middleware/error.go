package middleware

import (
	"errors"
	"net/http"

	"github.com/jessesomerville/yodahunters/internal/derror"
	"github.com/jessesomerville/yodahunters/internal/log"
)

// ErrorHandler handles responding to requests with error messages.
func ErrorHandler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			var serr *derror.ServerError
			if !errors.As(err, &serr) {
				serr = &derror.ServerError{Status: http.StatusInternalServerError, Err: err}
			}
			log.Errorf(r.Context(), "returning %d (%s) for error %v", serr.Status, http.StatusText(serr.Status), err)
			http.Error(w, serr.Err.Error(), serr.Status)
		}
	}
}
