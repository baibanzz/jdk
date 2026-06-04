package core

import "github.com/baibanzz/jdk/core/internal/jwt"

type TokenClaims[T any] = jwt.TokenClaims[T]

func NewJwt[T any](t T) *TokenClaims[T] {
	return jwt.New(t)
}
