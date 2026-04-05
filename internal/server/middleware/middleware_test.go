package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorizationHandler_ValidJWT(t *testing.T) {
	secret := []byte("12345678901234567890123456789012")
	jwt, err := GenerateJWT(42, true, secret)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	var capturedUserID int
	var capturedIsAdmin bool
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		capturedUserID = r.Context().Value(CtxUserKey).(int)
		capturedIsAdmin = r.Context().Value(CtxAdminKey).(bool)
	})

	handler := AuthorizationHandler(next, secret)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: jwt.Raw})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if capturedUserID != 42 {
		t.Errorf("got userID %d, want 42", capturedUserID)
	}
	if !capturedIsAdmin {
		t.Error("expected isAdmin to be true in context")
	}
}

func TestAuthorizationHandler_NoCookie(t *testing.T) {
	secret := []byte("12345678901234567890123456789012")
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := AuthorizationHandler(next, secret)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if called {
		t.Fatal("next handler should not have been called")
	}
	if rr.Code != http.StatusFound {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusFound)
	}
	loc := rr.Header().Get("Location")
	if loc != "/login" {
		t.Errorf("got redirect location %q, want /login", loc)
	}
}

func TestPageHandler_Defaults(t *testing.T) {
	var capturedPage any
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		capturedPage = r.Context().Value(CtxPageKey)
	})

	handler := PageHandler(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if capturedPage == nil {
		t.Fatal("page data was not set in context")
	}
}

func TestPageHandler_WithQueryParams(t *testing.T) {
	var capturedPage any

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPage = r.Context().Value(CtxPageKey)
	})

	handler := PageHandler(next)

	req := httptest.NewRequest(http.MethodGet, "/?page=3", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if capturedPage == nil {
		t.Fatal("page data was not set in context")
	}
}

func TestAdminHandler_IsAdmin(t *testing.T) {
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := AdminHandler(next)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := context.WithValue(req.Context(), CtxAdminKey, true)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("next handler was not called for admin user")
	}
}

func TestAdminHandler_NotAdmin(t *testing.T) {
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := AdminHandler(next)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := context.WithValue(req.Context(), CtxAdminKey, false)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if called {
		t.Fatal("next handler should not have been called for non-admin")
	}
	if rr.Code != http.StatusFound {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusFound)
	}
	loc := rr.Header().Get("Location")
	if loc != "/" {
		t.Errorf("got redirect location %q, want /", loc)
	}
}
