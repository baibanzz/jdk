package utils

import (
	"strings"

	"github.com/google/uuid"
)

func NewUUID() string {
	u := uuid.New().String()
	return strings.Replace(u, "-", "", -1)
}
