package utils

import (
	"encoding/hex"
	"testing"
)

func TestMD5(t *testing.T) {
	// 测试 string 输入
	strResult := MD5("hello")
	expected := "5d41402abc4b2a76b9719d911017c592"
	if hex.EncodeToString(strResult) != expected {
		t.Errorf("MD5(string) = %s, expected %s", hex.EncodeToString(strResult), expected)
	}

	// 测试 []byte 输入
	bytesResult := MD5([]byte("hello"))
	if hex.EncodeToString(bytesResult) != expected {
		t.Errorf("MD5([]byte) = %s, expected %s", hex.EncodeToString(bytesResult), expected)
	}

	// 测试空字符串
	emptyResult := MD5("")
	emptyExpected := "d41d8cd98f00b204e9800998ecf8427e"
	if hex.EncodeToString(emptyResult) != emptyExpected {
		t.Errorf("MD5(empty) = %s, expected %s", hex.EncodeToString(emptyResult), emptyExpected)
	}
}

func TestBASE64(t *testing.T) {
	// 测试 string 输入
	strResult := BASE64("hello")
	expected := "aGVsbG8="
	if strResult != expected {
		t.Errorf("BASE64(string) = %s, expected %s", strResult, expected)
	}

	// 测试 []byte 输入
	bytesResult := BASE64([]byte("hello"))
	if bytesResult != expected {
		t.Errorf("BASE64([]byte) = %s, expected %s", bytesResult, expected)
	}

	// 测试空字符串
	emptyResult := BASE64("")
	emptyExpected := ""
	if emptyResult != emptyExpected {
		t.Errorf("BASE64(empty) = %s, expected %s", emptyResult, emptyExpected)
	}
}

func TestBASE64Decode(t *testing.T) {
	// 正常解码
	data, err := BASE64Decode("aGVsbG8=")
	if err != nil {
		t.Errorf("BASE64Decode error: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("BASE64Decode = %s, expected hello", string(data))
	}

	// 解码空字符串
	emptyData, err := BASE64Decode("")
	if err != nil {
		t.Errorf("BASE64Decode(empty) error: %v", err)
	}
	if len(emptyData) != 0 {
		t.Errorf("BASE64Decode(empty) length = %d, expected 0", len(emptyData))
	}

	// 解码无效字符串
	_, err = BASE64Decode("invalid!!!")
	if err == nil {
		t.Errorf("BASE64Decode(invalid) should return error")
	}
}

func TestBASE64_RoundTrip(t *testing.T) {
	// 编码后再解码，应该得到原始数据
	original := "Hello, World! 你好世界"
	encoded := BASE64(original)
	t.Logf("Encoded: %s", encoded)
	decoded, err := BASE64Decode(encoded)
	if err != nil {
		t.Errorf("Round trip failed: %v", err)
	}
	if string(decoded) != original {
		t.Errorf("Round trip: got %s, expected %s", string(decoded), original)
	}
}
