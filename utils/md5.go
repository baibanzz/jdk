package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// MD5 加密
func MD5[T string | []byte](data T) string {
	m := md5.New()
	m.Write([]byte(data))
	passwordmdsBys := m.Sum(nil)
	return hex.EncodeToString(passwordmdsBys)
}

// SHA1加密
func SHA1[T string | []byte](str T) string {
	fmt.Println("str:", str)
	h := sha1.New()
	h.Write([]byte(str))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func HMACSHA1[T string | []byte](keyStr string, data T) string {
	//hmac ,use sha1
	key := []byte(keyStr)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(data))
	srcBytes := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(srcBytes)
}
