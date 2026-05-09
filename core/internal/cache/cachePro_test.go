package cache

import (
	"testing"
	"time"
)

// TestCacheProBasic 测试CachePro的基本功能
func TestCacheProBasic(t *testing.T) {
	tc := NewPro[interface{}](DefaultExpiration, 0, nil)

	// 测试Set和Get
	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", "hello", DefaultExpiration)
	tc.Set("c", 3.14, DefaultExpiration)

	// 测试Get
	a, found := tc.Get("a")
	if !found {
		t.Error("a was not found")
	}
	if a != 1 {
		t.Error("a is not 1")
	}

	b, found := tc.Get("b")
	if !found {
		t.Error("b was not found")
	}
	if b != "hello" {
		t.Error("b is not 'hello'")
	}

	c, found := tc.Get("c")
	if !found {
		t.Error("c was not found")
	}
	if c != 3.14 {
		t.Error("c is not 3.14")
	}

	// 测试Get不存在的键
	_, found = tc.Get("nonexistent")
	if found {
		t.Error("nonexistent key was found")
	}
}

// TestCacheProExpiration 测试CachePro的过期功能
func TestCacheProExpiration(t *testing.T) {
	tc := NewPro[int](50*time.Millisecond, 1*time.Millisecond, nil)

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", 2, NoExpiration)
	tc.Set("c", 3, 20*time.Millisecond)

	<-time.After(25 * time.Millisecond)
	_, found := tc.Get("c")
	if found {
		t.Error("Found c when it should have been automatically deleted")
	}

	<-time.After(30 * time.Millisecond)
	_, found = tc.Get("a")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}

	_, found = tc.Get("b")
	if !found {
		t.Error("Did not find b even though it was set to never expire")
	}
}

// TestCacheProCompute 测试计算函数
func TestCacheProCompute(t *testing.T) {
	tc := NewPro[int](DefaultExpiration, 0, nil)

	// 定义计算函数 - 加法
	addFunc := func(a, b int) int {
		return a + b
	}

	// 测试Compute - 键不存在时使用默认值
	result, err := tc.Compute("sum", addFunc, 10)
	if err != nil {
		t.Errorf("Compute failed: %v", err)
	}
	if result != 10 {
		t.Errorf("Expected 10, got %v", result)
	}

	// 测试Compute - 键存在时进行计算
	tc.Set("value", 5, DefaultExpiration)
	result, err = tc.Compute("value", addFunc, 0)
	if err != nil {
		t.Errorf("Compute failed: %v", err)
	}
	if result != 10 { // 5 + 5 = 10
		t.Errorf("Expected 10, got %v", result)
	}

	// 验证缓存中的值已更新
	value, found := tc.Get("value")
	if !found {
		t.Error("value was not found after compute")
	}
	if value != 10 {
		t.Errorf("Expected 10 in cache, got %v", value)
	}
}

// TestCacheProComputeWithExpiration 测试带过期时间的计算函数
func TestCacheProComputeWithExpiration(t *testing.T) {
	tc := NewPro[int](DefaultExpiration, 0, nil)

	// 定义计算函数 - 乘法
	multiplyFunc := func(a, b int) int {
		return a * b
	}

	// 测试ComputeWithExpiration
	result, err := tc.ComputeWithExpiration("product", multiplyFunc, 1, 50*time.Millisecond)
	if err != nil {
		t.Errorf("ComputeWithExpiration failed: %v", err)
	}
	if result != 1 {
		t.Errorf("Expected 1, got %v", result)
	}

	// 设置值并再次计算
	tc.Set("num", 3, DefaultExpiration)
	result, err = tc.ComputeWithExpiration("num", multiplyFunc, 0, 50*time.Millisecond)
	if err != nil {
		t.Errorf("ComputeWithExpiration failed: %v", err)
	}
	if result != 9 { // 3 * 3 = 9
		t.Errorf("Expected 9, got %v", result)
	}

	// 验证过期时间
	<-time.After(60 * time.Millisecond)
	_, found := tc.Get("num")
	if found {
		t.Error("num should have expired")
	}
}

// TestCacheProComputeTwoKeys 测试两个键的计算函数
func TestCacheProComputeTwoKeys(t *testing.T) {
	tc := NewPro[string](DefaultExpiration, 0, nil)

	// 定义计算函数 - 字符串连接
	concatFunc := func(a, b string) string {
		return a + " " + b
	}

	// 设置两个键
	tc.Set("first", "Hello", DefaultExpiration)
	tc.Set("second", "World", DefaultExpiration)

	// 测试ComputeTwoKeys
	result, err := tc.ComputeTwoKeys("first", "second", concatFunc, "message", DefaultExpiration)
	if err != nil {
		t.Errorf("ComputeTwoKeys failed: %v", err)
	}
	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%v'", result)
	}

	// 验证结果已存储
	message, found := tc.Get("message")
	if !found {
		t.Error("message was not found")
	}
	if message != "Hello World" {
		t.Errorf("Expected 'Hello World' in cache, got '%v'", message)
	}
}

// TestCacheProComputeTwoKeysError 测试两个键计算函数的错误情况
func TestCacheProComputeTwoKeysError(t *testing.T) {
	tc := NewPro[int](DefaultExpiration, 0, nil)

	// 定义计算函数
	addFunc := func(a, b int) int {
		return a + b
	}

	// 测试键不存在的情况
	_, err := tc.ComputeTwoKeys("nonexistent1", "nonexistent2", addFunc, "result", DefaultExpiration)
	if err == nil {
		t.Error("Expected error for nonexistent keys")
	}

	// 测试一个键存在，一个键不存在的情况
	tc.Set("exists", 5, DefaultExpiration)
	_, err = tc.ComputeTwoKeys("exists", "nonexistent", addFunc, "result", DefaultExpiration)
	if err == nil {
		t.Error("Expected error for one nonexistent key")
	}
}

// TestCacheProDelete 测试删除功能
func TestCacheProDelete(t *testing.T) {
	tc := NewPro[string](DefaultExpiration, 0, nil)

	tc.Set("key", "value", DefaultExpiration)
	tc.Delete("key")

	_, found := tc.Get("key")
	if found {
		t.Error("key was found after deletion")
	}
}

// TestCacheProFlush 测试清空功能
func TestCacheProFlush(t *testing.T) {
	tc := NewPro[int](DefaultExpiration, 0, nil)

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", 2, DefaultExpiration)
	tc.Set("c", 3, DefaultExpiration)

	if tc.ItemCount() != 3 {
		t.Errorf("Expected 3 items, got %d", tc.ItemCount())
	}

	tc.Flush()

	if tc.ItemCount() != 0 {
		t.Errorf("Expected 0 items after flush, got %d", tc.ItemCount())
	}
}

// TestCacheProItemCount 测试项目计数
func TestCacheProItemCount(t *testing.T) {
	tc := NewPro[int](DefaultExpiration, 0, nil)

	if tc.ItemCount() != 0 {
		t.Errorf("Expected 0 items, got %d", tc.ItemCount())
	}

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", 2, DefaultExpiration)

	if tc.ItemCount() != 2 {
		t.Errorf("Expected 2 items, got %d", tc.ItemCount())
	}
}

// TestCacheProWithStruct 测试使用结构体
func TestCacheProWithStruct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	tc := NewPro[Person](DefaultExpiration, 0, nil)

	person := Person{Name: "Alice", Age: 30}
	tc.Set("person", person, DefaultExpiration)

	result, found := tc.Get("person")
	if !found {
		t.Error("person was not found")
	}

	if result.Name != "Alice" || result.Age != 30 {
		t.Errorf("Expected person {Alice 30}, got %+v", result)
	}
}
