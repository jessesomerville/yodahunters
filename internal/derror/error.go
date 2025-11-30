// Package derror provides common error handling types.
package derror

import (
	"fmt"
	"net/http"
)

// ServerError is an error to return in the server response.
type ServerError struct {
	Status int
	Err    error
}

func (s *ServerError) Error() string {
	return fmt.Sprintf("%d (%s): %v", s.Status, http.StatusText(s.Status), s.Err)
}
