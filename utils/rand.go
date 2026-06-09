package utils

import (
	"math/rand"
	"time"
)

// 预定义字符集常量
const (
	CHARSET_LOWER    = "abcdefghijklmnopqrstuvwxyz"  // 小写字母
	CHARSET_UPPER    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"  // 大写字母
	CHARSET_NUM      = "0123456789"                  // 数字
	CHARSET_SPECIAL  = "!@#$%^&*()_+-=[]{}|;:',.<>?" // 特殊符号
	CHARSET_ALPHA    = CHARSET_LOWER + CHARSET_UPPER // 字母
	CHARSET_ALPHANUM = CHARSET_ALPHA + CHARSET_NUM   // 字母+数字
)

// RandNum 生成指定长度的纯数字随机数（返回 int）
// 第一位不能为0，其后各位可以为0
func RandNum(length int) int {
	if length <= 0 {
		return 0
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	if length == 1 {
		return 1 + r.Intn(9) // 1-9
	}

	// 计算范围：[10^(length-1), 10^length)
	min := 1
	for i := 1; i < length; i++ {
		min *= 10
	}
	max := min * 10

	return min + r.Intn(max-min)
}

// RandString 使用指定字符集生成随机字符串，默认使用小写字母
// length: 字符串长度
// chars: 可选参数，指定字符集，默认为 CHARSET_LOWER
func RandString(length int, chars ...string) string {
	if length <= 0 {
		return ""
	}

	charset := CHARSET_LOWER
	if len(chars) > 0 && len(chars[0]) > 0 {
		for _, c := range chars {
			charset += c
		}
	}

	bytes := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	charsetLen := len(charset)
	for i := 0; i < length; i++ {
		bytes[i] = charset[r.Intn(charsetLen)]
	}
	return string(bytes)
}
