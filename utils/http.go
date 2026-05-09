package utils

import "net/http"

type Http struct {
	http.Client
}

func NewHttp() *Http {
	return &Http{http.Client{}}
}

//func (h *Http) Get[T any](url string) (data map[string]any, error) {
//
//}
