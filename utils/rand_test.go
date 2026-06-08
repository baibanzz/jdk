package utils

import (
	"fmt"
	"testing"
)

func TestRandNum(t *testing.T) {
	// 测试不同长度
	testCases := []struct {
		length int
	}{
		{1},
		{2},
		{3},
		{4},
		{6},
		{10},
	}

	for range testCases {
		for _, tc := range testCases {
			result := RandNum(tc.length)
			fmt.Printf("RandNum(%d) = %d\n", tc.length, result)

			// 验证返回类型
			if result <= 0 {
				t.Errorf("RandNum(%d) 返回值应大于0，实际为 %d", tc.length, result)
			}
		}
	}
}

func TestRandString(t *testing.T) {
	// 测试默认字符集（小写字母）
	result1 := RandString(6)
	fmt.Printf("RandString(6) = %s\n", result1)
	if len(result1) != 6 {
		t.Errorf("RandString(6) 长度应为6，实际为 %d", len(result1))
	}

	// 测试指定字符集
	result2 := RandString(6, CHARSET_UPPER)
	fmt.Printf("RandString(6, CHARSET_UPPER) = %s\n", result2)
	if len(result2) != 6 {
		t.Errorf("RandString(6, CHARSET_UPPER) 长度应为6，实际为 %d", len(result2))
	}

	// 测试数字字符集
	result3 := RandString(6, CHARSET_NUM)
	fmt.Printf("RandString(6, CHARSET_NUM) = %s\n", result3)
	if len(result3) != 6 {
		t.Errorf("RandString(6, CHARSET_NUM) 长度应为6，实际为 %d", len(result3))
	}

	// 测试字母+数字字符集
	result4 := RandString(8, CHARSET_ALPHANUM)
	fmt.Printf("RandString(8, CHARSET_ALPHANUM) = %s\n", result4)
	if len(result4) != 8 {
		t.Errorf("RandString(8, CHARSET_ALPHANUM) 长度应为8，实际为 %d", len(result4))
	}
}

func TestCharsetConstants(t *testing.T) {
	fmt.Printf("CHARSET_LOWER: %s\n", CHARSET_LOWER)
	fmt.Printf("CHARSET_UPPER: %s\n", CHARSET_UPPER)
	fmt.Printf("CHARSET_NUM: %s\n", CHARSET_NUM)
	fmt.Printf("CHARSET_SPECIAL: %s\n", CHARSET_SPECIAL)
	fmt.Printf("CHARSET_ALPHA: %s\n", CHARSET_ALPHA)
	fmt.Printf("CHARSET_ALPHANUM: %s\n", CHARSET_ALPHANUM)

	// 验证字符集长度
	if len(CHARSET_LOWER) != 26 {
		t.Errorf("CHARSET_LOWER 长度应为26，实际为 %d", len(CHARSET_LOWER))
	}
	if len(CHARSET_UPPER) != 26 {
		t.Errorf("CHARSET_UPPER 长度应为26，实际为 %d", len(CHARSET_UPPER))
	}
	if len(CHARSET_NUM) != 10 {
		t.Errorf("CHARSET_NUM 长度应为10，实际为 %d", len(CHARSET_NUM))
	}
	if len(CHARSET_ALPHA) != 52 {
		t.Errorf("CHARSET_ALPHA 长度应为52，实际为 %d", len(CHARSET_ALPHA))
	}
	if len(CHARSET_ALPHANUM) != 62 {
		t.Errorf("CHARSET_ALPHANUM 长度应为62，实际为 %d", len(CHARSET_ALPHANUM))
	}
}

func BenchmarkRandNum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandNum(6)
	}
}

func BenchmarkRandString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandString(6)
	}
}

func BenchmarkRandStringWithCharset(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandString(6, CHARSET_ALPHANUM)
	}
}
