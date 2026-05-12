package jwt

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type TokenClaims[T any] struct {
	Claims    T
	ExpiresAt int64 `json:"exp"`
}

func (t *TokenClaims[T]) SignToken(secret string) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	message := encodedHeader + "." + encodedPayload
	signature := hmacSHA256(secret, message)
	return message + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (t *TokenClaims[T]) ParseToken(secret, token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ErrInvalidToken
	}
	message := parts[0] + "." + parts[1]
	expected := hmacSHA256(secret, message)
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(expected, actual) {
		return ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ErrInvalidToken
	}
	if err = json.Unmarshal(payload, t); err != nil {
		return ErrInvalidToken
	}
	if t.ExpiresAt == 0 {
		return nil
	}
	if t.ExpiresAt <= time.Now().Unix() {
		return ErrExpiredToken
	}
	return nil
}

func hmacSHA256(secret, message string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

func NewTokenID(prefix string) string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return prefix + "_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}
