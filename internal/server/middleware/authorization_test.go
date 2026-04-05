package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorize_ValidCookie(t *testing.T) {
	secret := []byte("12345678901234567890123456789012")
	jwt, err := GenerateJWT(42, false, secret)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: jwt.Raw})

	userID, err := Authorize(req, secret)
	if err != nil {
		t.Fatalf("Authorize returned error: %v", err)
	}
	if userID != 42 {
		t.Errorf("got userID %d, want 42", userID)
	}
}

func TestAuthorize_NoCookie(t *testing.T) {
	secret := []byte("12345678901234567890123456789012")
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := Authorize(req, secret)
	if err == nil {
		t.Fatal("expected error when no cookie present")
	}
}

func TestAuthorize_InvalidJWTString(t *testing.T) {
	secret := []byte("12345678901234567890123456789012")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "not-a-valid-jwt"})

	_, err := Authorize(req, secret)
	if err == nil {
		t.Fatal("expected error for invalid JWT string")
	}
}

func TestAuthorize_WrongSecret(t *testing.T) {
	signingSecret := []byte("12345678901234567890123456789012")
	wrongSecret := []byte("abcdefghijklmnopqrstuvwxyz123456")

	jwt, err := GenerateJWT(7, false, signingSecret)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: jwt.Raw})

	_, err = Authorize(req, wrongSecret)
	if err == nil {
		t.Fatal("expected error when verifying with wrong secret")
	}
}

func TestIsAdmin_AdminUser(t *testing.T) {
	secret := []byte("12345678901234567890123456789012")
	jwt, err := GenerateJWT(1, true, secret)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: jwt.Raw})

	isAdmin, err := IsAdmin(req, secret)
	if err != nil {
		t.Fatalf("IsAdmin returned error: %v", err)
	}
	if !isAdmin {
		t.Error("expected isAdmin to be true")
	}
}

func TestIsAdmin_NonAdminUser(t *testing.T) {
	secret := []byte("12345678901234567890123456789012")
	jwt, err := GenerateJWT(1, false, secret)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: jwt.Raw})

	isAdmin, err := IsAdmin(req, secret)
	if err != nil {
		t.Fatalf("IsAdmin returned error: %v", err)
	}
	if isAdmin {
		t.Error("expected isAdmin to be false")
	}
}

func TestIsAdmin_NoCookie(t *testing.T) {
	secret := []byte("12345678901234567890123456789012")
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := IsAdmin(req, secret)
	if err == nil {
		t.Fatal("expected error when no cookie present")
	}
}
