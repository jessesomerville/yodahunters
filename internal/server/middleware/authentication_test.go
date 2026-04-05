package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-testing!")

	tests := []struct {
		name    string
		userID  int
		isAdmin bool
	}{
		{name: "regular user", userID: 1, isAdmin: false},
		{name: "admin user", userID: 42, isAdmin: true},
		{name: "zero user id", userID: 0, isAdmin: false},
		{name: "large user id", userID: 999999, isAdmin: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwt, err := GenerateJWT(tt.userID, tt.isAdmin, secret)
			if err != nil {
				t.Fatalf("GenerateJWT() returned error: %v", err)
			}

			if jwt.Header.Alg != "HS256" {
				t.Errorf("Header.Alg = %q, want %q", jwt.Header.Alg, "HS256")
			}
			if jwt.Header.Typ != "JWT" {
				t.Errorf("Header.Typ = %q, want %q", jwt.Header.Typ, "JWT")
			}
			if jwt.Payload.UserID != tt.userID {
				t.Errorf("Payload.UserID = %d, want %d", jwt.Payload.UserID, tt.userID)
			}
			if jwt.Payload.IsAdmin != tt.isAdmin {
				t.Errorf("Payload.IsAdmin = %v, want %v", jwt.Payload.IsAdmin, tt.isAdmin)
			}

			// Exp should be roughly 12 hours from now.
			expectedExp := int(time.Now().Add(12 * time.Hour).Unix())
			if abs(jwt.Payload.Exp-expectedExp) > 5 {
				t.Errorf("Payload.Exp = %d, want approximately %d", jwt.Payload.Exp, expectedExp)
			}

			if jwt.Raw == "" {
				t.Error("Raw is empty")
			}
			if len(jwt.Signature) == 0 {
				t.Error("Signature is empty")
			}

			// Roundtrip through ParseJWT.
			parsed, err := ParseJWT(jwt.Raw)
			if err != nil {
				t.Fatalf("ParseJWT() roundtrip returned error: %v", err)
			}
			if parsed.Payload.UserID != tt.userID {
				t.Errorf("roundtrip Payload.UserID = %d, want %d", parsed.Payload.UserID, tt.userID)
			}
			if parsed.Payload.IsAdmin != tt.isAdmin {
				t.Errorf("roundtrip Payload.IsAdmin = %v, want %v", parsed.Payload.IsAdmin, tt.isAdmin)
			}
			if parsed.Header.Alg != "HS256" {
				t.Errorf("roundtrip Header.Alg = %q, want %q", parsed.Header.Alg, "HS256")
			}
		})
	}
}

func TestParseJWT(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-testing!")

	// Generate a valid token to use as the base valid case.
	validJWT, err := GenerateJWT(7, false, secret)
	if err != nil {
		t.Fatalf("setup: GenerateJWT() returned error: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid JWT",
			input:   validJWT.Raw,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "one part",
			input:   "onlyonepart",
			wantErr: true,
		},
		{
			name:    "two parts",
			input:   "part1.part2",
			wantErr: true,
		},
		{
			name:    "four parts",
			input:   "a.b.c.d",
			wantErr: true,
		},
		{
			name:    "invalid base64 in header",
			input:   "!!!.dGVzdA.dGVzdA",
			wantErr: true,
		},
		{
			name:    "invalid base64 in payload",
			input:   base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`)) + ".!!!.dGVzdA",
			wantErr: true,
		},
		{
			name:    "invalid JSON in header",
			input:   base64.RawURLEncoding.EncodeToString([]byte(`not json`)) + "." + base64.RawURLEncoding.EncodeToString([]byte(`{"user_id":1,"is_admin":false,"exp":999}`)) + ".dGVzdA",
			wantErr: true,
		},
		{
			name:    "invalid JSON in payload",
			input:   base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte(`not json`)) + ".dGVzdA",
			wantErr: true,
		},
		{
			name:    "invalid base64 in signature",
			input:   base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte(`{"user_id":1,"is_admin":false,"exp":999}`)) + ".!!!",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseJWT(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("ParseJWT() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseJWT() unexpected error: %v", err)
			}
			if parsed.Payload.UserID != 7 {
				t.Errorf("Payload.UserID = %d, want 7", parsed.Payload.UserID)
			}
			if parsed.Raw != tt.input {
				t.Errorf("Raw = %q, want %q", parsed.Raw, tt.input)
			}
		})
	}
}

func TestJWTIsValid(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-testing!")
	wrongSecret := []byte("wrong-secret-key-for-testing!!!!")

	tests := []struct {
		name      string
		setupJWT  func(t *testing.T) JWT
		secret    []byte
		wantValid bool
	}{
		{
			name: "valid signature and not expired",
			setupJWT: func(t *testing.T) JWT {
				t.Helper()
				jwt, err := GenerateJWT(1, false, secret)
				if err != nil {
					t.Fatalf("GenerateJWT() returned error: %v", err)
				}
				return jwt
			},
			secret:    secret,
			wantValid: true,
		},
		{
			name: "wrong secret",
			setupJWT: func(t *testing.T) JWT {
				t.Helper()
				jwt, err := GenerateJWT(1, false, secret)
				if err != nil {
					t.Fatalf("GenerateJWT() returned error: %v", err)
				}
				return jwt
			},
			secret:    wrongSecret,
			wantValid: false,
		},
		{
			name: "expired token",
			setupJWT: func(t *testing.T) JWT {
				t.Helper()
				return buildExpiredJWT(t, 1, false, secret)
			},
			secret:    secret,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwt := tt.setupJWT(t)
			valid, err := jwt.IsValid(tt.secret)
			if err != nil {
				t.Fatalf("IsValid() returned error: %v", err)
			}
			if valid != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestJWTString(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-testing!")

	jwt, err := GenerateJWT(5, true, secret)
	if err != nil {
		t.Fatalf("GenerateJWT() returned error: %v", err)
	}

	if jwt.String() != jwt.Raw {
		t.Errorf("String() = %q, want %q", jwt.String(), jwt.Raw)
	}

	// Verify the string has three dot-separated parts.
	parts := strings.Split(jwt.String(), ".")
	if len(parts) != 3 {
		t.Errorf("String() has %d parts, want 3", len(parts))
	}
}

// buildExpiredJWT manually constructs a JWT with an expiration in the past,
// properly signed with the given secret.
func buildExpiredJWT(t *testing.T, userID int, isAdmin bool, secret []byte) JWT {
	t.Helper()

	header := joseHeader{Alg: "HS256", Typ: "JWT"}
	payload := jwsPayload{
		UserID:  userID,
		IsAdmin: isAdmin,
		Exp:     int(time.Now().Add(-1 * time.Hour).Unix()), // expired 1 hour ago
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("json.Marshal(header) returned error: %v", err)
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal(payload) returned error: %v", err)
	}

	message := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(message))
	sig := h.Sum(nil)

	raw := message + "." + base64.RawURLEncoding.EncodeToString(sig)

	return JWT{
		Header:    header,
		Payload:   payload,
		Signature: sig,
		Raw:       raw,
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
