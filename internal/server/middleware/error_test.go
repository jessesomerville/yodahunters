package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jessesomerville/yodahunters/internal/derror"
)

func TestErrorHandler(t *testing.T) {
	tests := []struct {
		name       string
		handler    func(w http.ResponseWriter, r *http.Request) error
		wantStatus int
		wantBody   string
	}{
		{
			name: "handler returns nil gives 200 OK",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				w.WriteHeader(http.StatusOK)
				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "handler returns plain error gives 500",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errors.New("something broke")
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "something broke",
		},
		{
			name: "handler returns ServerError 404",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return &derror.ServerError{Status: http.StatusNotFound, Err: errors.New("not found")}
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name: "handler returns ServerError 403",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return &derror.ServerError{Status: http.StatusForbidden, Err: errors.New("forbidden")}
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ErrorHandler(tt.handler)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test", nil)

			h.ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.wantBody {
					t.Errorf("body = %q, want %q", body, tt.wantBody)
				}
			}
		})
	}
}
