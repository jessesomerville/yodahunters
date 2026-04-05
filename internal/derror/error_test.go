package derror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestServerError_Error(t *testing.T) {
	tests := []struct {
		name   string
		status int
		err    error
		want   string
	}{
		{
			name:   "not found",
			status: http.StatusNotFound,
			err:    errors.New("page missing"),
			want:   fmt.Sprintf("404 (%s): page missing", http.StatusText(http.StatusNotFound)),
		},
		{
			name:   "internal server error",
			status: http.StatusInternalServerError,
			err:    errors.New("something broke"),
			want:   fmt.Sprintf("500 (%s): something broke", http.StatusText(http.StatusInternalServerError)),
		},
		{
			name:   "bad request",
			status: http.StatusBadRequest,
			err:    errors.New("invalid input"),
			want:   fmt.Sprintf("400 (%s): invalid input", http.StatusText(http.StatusBadRequest)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se := &ServerError{Status: tt.status, Err: tt.err}
			if got := se.Error(); got != tt.want {
				t.Errorf("ServerError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestServerError_ImplementsError(t *testing.T) {
	var _ error = (*ServerError)(nil)
}
