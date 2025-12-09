package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type joseHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwsPayload struct {
	UserID int `json:"user_id"`
	Exp    int `json:"exp"`
}

// JWT is a struct that holds the relevant data for handling JWTs.
type JWT struct {
	Header    joseHeader
	Payload   jwsPayload
	Signature []byte
	Raw       string
}

// GenerateJWT takes a user id and the signing secret and generates a JWT.
// with the following structure:
//  Header: {"alg":"HS256", "typ":"JWT"}
//  Claims: {"user_id": user_id, "exp": [current time + 12hrs]}
func GenerateJWT(userID int, secret []byte) (JWT, error) {
	// Set the header and payload
	jwt := JWT{
		Header: joseHeader{
			Alg: "HS256",
			Typ: "JWT",
		},
		Payload: jwsPayload{
			UserID: userID,
			Exp:    int((time.Now().Add(12 * time.Hour)).Unix()),
		},
		Signature: nil,
		Raw:       "",
	}

	// Sign the JWT
	h := hmac.New(sha256.New, secret)
	headerJSON, err := json.Marshal(jwt.Header)
	if err != nil {
		return JWT{}, err
	}
	payloadJSON, err := json.Marshal(jwt.Payload)
	if err != nil {
		return JWT{}, err
	}

	message := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	h.Write([]byte(message))
	jwt.Signature = h.Sum(nil)
	jwt.Raw = message + "." + base64.RawURLEncoding.EncodeToString(jwt.Signature)
	return jwt, nil
}

// ParseJWT parses a JWT in string form, and returns a JWT struct
// with the fields filled out appropriately.
func ParseJWT(s string) (JWT, error) {
	jwtParts := strings.Split(s, ".")
  if len(jwtParts) != 3 {
    return JWT{}, errors.New("malformed JWT")
  }
	headerBytes, err := base64.RawURLEncoding.DecodeString(jwtParts[0])
	if err != nil {
		return JWT{}, err
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(jwtParts[1])
	if err != nil {
		return JWT{}, err
	}

	var header joseHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return JWT{}, err
	}
	var payload jwsPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return JWT{}, err
	}

	signature, err := base64.RawURLEncoding.DecodeString(jwtParts[2])
	if err != nil {
		return JWT{}, err
	}

	jwt := JWT{
		Header:    header,
		Payload:   payload,
		Signature: signature,
		Raw:       s,
	}

	return jwt, nil
}

// String returns a string with the encoded JWT.
func (j *JWT) String() string {
	return j.Raw
}

// IsValid verifies that the JWT data supplied in the struct is
// was signed by the secret supplied, and that the JWT has not expired.
func (j *JWT) IsValid(secret []byte) (bool, error) {
	signatureStart := strings.LastIndex(j.Raw, ".")
	headerPayload := j.Raw[:signatureStart]

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(headerPayload))
	if bytes.Equal(j.Signature, h.Sum(nil)) && (int(time.Now().Unix()) < j.Payload.Exp) {
		return true, nil
	}
	return false, nil
}

// GeneratePasswordHash adds a hashed password to a User struct if  there is a
// password in the struct, and a password hash is not already present.
func (u *User) GeneratePasswordHash() error {
	// If no password is provided, then it should be set to an empty string.
	// It shouldn't be possible to set an empty string as a password in any case.
	if u.PasswordHash != nil {
		return errors.New("user struct already contains password hash")
	}
	if u.Password == "" {
		return errors.New("attempted to generate password hash where password is an empty string")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	u.Password = ""
	return nil
}
