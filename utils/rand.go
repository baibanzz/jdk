package utils

import (
	"math/rand"
	"strings"
	"time"
)

func RandString(len int) string {

	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
		//time.Sleep(1)
	}
	return strings.ToLower(string(bytes))
}
