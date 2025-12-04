package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

type joseHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwsPayload struct {
	UserID int `json:"user_id"`
	Exp    int `json:"exp"`
}

type JWT struct {
	Header    joseHeader
	Payload   jwsPayload
	Signature []byte
	Raw       string
}

func GenerateJWT(user_id int, secret []byte) (JWT, error) {
	// Set the header and payload
	jwt := JWT{
		Header: joseHeader{
			Alg: "HS256",
			Typ: "JWT",
		},
		Payload: jwsPayload{
			UserID: user_id,
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

func ParseJWT(s string) (JWT, error) {
	jwtParts := strings.Split(s, ".")

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

func (j *JWT) String() string {
	return j.Raw
}

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
