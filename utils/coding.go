package utils

import (
	"crypto/md5"
	"encoding/base64"
)

func MD5[T string | []byte](data T) []byte {
	m := md5.New()
	switch v := any(data).(type) {
	case string:
		m.Write([]byte(v))
	case []byte:
		m.Write(v)
	}
	return m.Sum(nil)
}

// BASE64 编码，支持 string 或 []byte
func BASE64[T string | []byte](data T) string {
	var b []byte
	switch v := any(data).(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	}
	return base64.StdEncoding.EncodeToString(b)
}

// BASE64Decode Base64 解码，返回 []byte 和错误
func BASE64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
