package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jessesomerville/yodahunters/internal/server/middleware"
)

func TestAPIHandleGetThreadByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	wantThread := Thread{ID: 5, Title: "hello", Body: "world", AuthorID: 1, CreatedAt: fixedTime}

	tests := []struct {
		name     string
		pathID   string
		store    *fakeStore
		wantCode int
		wantBody *Thread // nil means don't verify body
	}{
		{
			name:   "returns thread for valid id",
			pathID: "5",
			store: &fakeStore{
				GetThreadByIDFn: func(_ context.Context, id int) (Thread, error) {
					return wantThread, nil
				},
			},
			wantCode: http.StatusOK,
			wantBody: &wantThread,
		},
		{
			name:     "non-numeric id returns 500",
			pathID:   "notanint",
			store:    &fakeStore{},
			wantCode: http.StatusInternalServerError,
		},
		{
			name:   "store error returns 500",
			pathID: "5",
			store: &fakeStore{
				GetThreadByIDFn: func(_ context.Context, _ int) (Thread, error) {
					return Thread{}, errors.New("db unavailable")
				},
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{store: tt.store}

			r := newRequest(http.MethodGet, "/threads/"+tt.pathID, nil, 1)
			r.SetPathValue("id", tt.pathID)
			w := httptest.NewRecorder()

			middleware.ErrorHandler(s.apiHandleGetThreadByID).ServeHTTP(w, r)

			if w.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", w.Code, tt.wantCode)
			}
			if tt.wantBody != nil {
				var got Thread
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatalf("decode response body: %v", err)
				}
				if got != *tt.wantBody {
					t.Errorf("body = %+v, want %+v", got, *tt.wantBody)
				}
			}
		})
	}
}
