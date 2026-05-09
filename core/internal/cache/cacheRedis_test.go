package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	addrhost = "127.0.0.1:6379"
	password = "88888888"
)

// TestCacheRedisBasic 测试CacheRedis的基本功能
// 注意：这个测试需要运行Redis服务器
func TestCacheRedisBasic(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis服务器未运行，跳过测试")
	}

	// 清空数据库
	client.FlushDB(ctx)

	// 创建CacheRedis实例
	cache := NewCacheRedis[string](client, 5*time.Minute, 5*time.Minute)

	// 测试Set和Get
	err := cache.Set("key1", "value1", DefaultExpiration)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	value, found := cache.Get("key1")
	if !found {
		t.Error("Get failed: key1 not found")
	}
	if value != "value1" {
		t.Errorf("Get returned wrong value: got %v, want %v", value, "value1")
	}

	// 测试Get不存在的键
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Get should not find nonexistent key")
	}

	// 测试Delete
	err = cache.Delete("key1")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	_, found = cache.Get("key1")
	if found {
		t.Error("Key should be deleted")
	}

	// 清理
	client.FlushDB(ctx)
	client.Close()
}

// TestCacheRedisExpiration 测试过期功能
func TestCacheRedisExpiration(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis服务器未运行，跳过测试")
	}

	// 清空数据库
	client.FlushDB(ctx)

	// 创建CacheRedis实例
	cache := NewCacheRedis[string](client, 2*time.Second, 2*time.Second)

	// 测试带过期时间的Set
	err := cache.Set("key2", "value2", 2*time.Second)
	if err != nil {
		t.Errorf("Set with expiration failed: %v", err)
	}

	// 立即获取应该存在
	value, found := cache.Get("key2")
	if !found {
		t.Error("Key should exist immediately after set")
	}
	if value != "value2" {
		t.Errorf("Get returned wrong value: got %v, want %v", value, "value2")
	}

	// 等待过期
	time.Sleep(10 * time.Second)

	// 过期后应该不存在
	_, found = cache.Get("key2")
	if found {
		t.Error("Key should be expired")
	}

	// 清理
	client.FlushDB(ctx)
	client.Close()
}

// TestCacheRedisAdd 测试Add功能（仅当键不存在时添加）
func TestCacheRedisAdd(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis服务器未运行，跳过测试")
	}

	// 清空数据库
	client.FlushDB(ctx)

	// 创建CacheRedis实例
	cache := NewCacheRedis[string](client, 5*time.Minute, 5*time.Minute)

	// 第一次Add应该成功
	err := cache.Add("key3", "value3", DefaultExpiration)
	if err != nil {
		t.Errorf("First Add failed: %v", err)
	}

	// 第二次Add应该失败
	err = cache.Add("key3", "value4", DefaultExpiration)
	if err == nil {
		t.Error("Second Add should fail")
	}
	if err.Error() != "item key3 already exists" {
		t.Errorf("Wrong error message: got %v, want %v", err.Error(), "item key3 already exists")
	}

	// 验证值仍然是第一次设置的
	value, found := cache.Get("key3")
	if !found {
		t.Error("Key should exist")
	}
	if value != "value3" {
		t.Errorf("Get returned wrong value: got %v, want %v", value, "value3")
	}

	// 清理
	client.FlushDB(ctx)
	client.Close()
}

// TestCacheRedisReplace 测试Replace功能（仅当键存在时替换）
func TestCacheRedisReplace(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis服务器未运行，跳过测试")
	}

	// 清空数据库
	client.FlushDB(ctx)

	// 创建CacheRedis实例
	cache := NewCacheRedis[string](client, 5*time.Minute, 5*time.Minute)

	// 替换不存在的键应该失败
	err := cache.Replace("key4", "value4", DefaultExpiration)
	if err == nil {
		t.Error("Replace on non-existent key should fail")
	}

	// 先设置键
	err = cache.Set("key4", "oldvalue", DefaultExpiration)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// 替换存在的键应该成功
	err = cache.Replace("key4", "newvalue", DefaultExpiration)
	if err != nil {
		t.Errorf("Replace failed: %v", err)
	}

	// 验证值已被替换
	value, found := cache.Get("key4")
	if !found {
		t.Error("Key should exist")
	}
	if value != "newvalue" {
		t.Errorf("Get returned wrong value: got %v, want %v", value, "newvalue")
	}

	// 清理
	client.FlushDB(ctx)
	client.Close()
}

// TestCacheRedisIncrementDecrement 测试增加和减少功能
func TestCacheRedisIncrementDecrement(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis服务器未运行，跳过测试")
	}

	// 清空数据库
	client.FlushDB(ctx)

	// 创建CacheRedis实例
	cache := NewCacheRedis[string](client, 5*time.Minute, 5*time.Minute)

	// 测试Increment
	// 注意：Increment/Decrement方法只适用于数值类型
	// 这里我们测试int64类型
	// 首先设置一个数值
	err := cache.Set("counter", "10", DefaultExpiration)
	if err != nil {
		t.Errorf("Set counter failed: %v", err)
	}

	// 由于我们的泛型实现，Increment/Decrement方法需要数值类型
	// 这里我们创建一个专门用于int的缓存实例
	intCache := NewCacheRedis[int64](client, 5*time.Minute, 5*time.Minute)
	intCache.Set("intcounter", int64(10), DefaultExpiration)

	// 测试Increment
	result, err := intCache.Increment("intcounter", 5)
	if err != nil {
		t.Errorf("Increment failed: %v", err)
	}
	if result != 15 {
		t.Errorf("Increment wrong result: got %v, want %v", result, 15)
	}

	// 测试Decrement
	result, err = intCache.Decrement("intcounter", 3)
	if err != nil {
		t.Errorf("Decrement failed: %v", err)
	}
	if result != 12 {
		t.Errorf("Decrement wrong result: got %v, want %v", result, 12)
	}

	// 清理
	client.FlushDB(ctx)
	client.Close()
}

// TestCacheRedisFlush 测试清空功能
func TestCacheRedisFlush(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis服务器未运行，跳过测试")
	}

	// 清空数据库
	client.FlushDB(ctx)

	// 创建CacheRedis实例
	cache := NewCacheRedis[string](client, 5*time.Minute, 5*time.Minute)

	// 设置一些键
	cache.Set("key5", "value5", DefaultExpiration)
	cache.Set("key6", "value6", DefaultExpiration)
	cache.Set("key7", "value7", DefaultExpiration)

	// 验证键存在
	if count, err := cache.ItemCount(); err != nil || count != 3 {
		t.Errorf("ItemCount wrong: got %v, want %v, error: %v", count, 3, err)
	}

	// 清空缓存
	err := cache.Flush()
	if err != nil {
		t.Errorf("Flush failed: %v", err)
	}

	// 验证所有键都被清空
	if count, err := cache.ItemCount(); err != nil || count != 0 {
		t.Errorf("After Flush, ItemCount wrong: got %v, want %v, error: %v", count, 0, err)
	}

	// 清理
	client.Close()
}

// TestCacheRedisStruct 测试结构体存储
func TestCacheRedisStruct(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis服务器未运行，跳过测试")
	}

	// 清空数据库
	client.FlushDB(ctx)

	// 定义测试结构体
	type Person struct {
		Name string
		Age  int
	}

	// 创建CacheRedis实例
	cache := NewCacheRedis[Person](client, 5*time.Minute, 5*time.Minute)

	// 设置结构体
	person := Person{Name: "Alice", Age: 30}
	err := cache.Set("person1", person, DefaultExpiration)
	if err != nil {
		t.Errorf("Set struct failed: %v", err)
	}

	// 获取结构体
	retrieved, found := cache.Get("person1")
	if !found {
		t.Error("Struct not found")
	}
	if retrieved.Name != "Alice" || retrieved.Age != 30 {
		t.Errorf("Struct mismatch: got %+v, want %+v", retrieved, person)
	}

	// 清理
	client.FlushDB(ctx)
	client.Close()
}

// BenchmarkCacheRedisSet 基准测试Set操作
func BenchmarkCacheRedisSet(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skip("Redis服务器未运行，跳过基准测试")
	}

	client.FlushDB(ctx)
	cache := NewCacheRedis[string](client, 5*time.Minute, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, "value", DefaultExpiration)
	}

	b.StopTimer()
	client.FlushDB(ctx)
	client.Close()
}

// BenchmarkCacheRedisGet 基准测试Get操作
func BenchmarkCacheRedisGet(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skip("Redis服务器未运行，跳过基准测试")
	}

	client.FlushDB(ctx)
	cache := NewCacheRedis[string](client, 5*time.Minute, 5*time.Minute)

	// 先设置一个键
	cache.Set("benchkey", "benchvalue", DefaultExpiration)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("benchkey")
	}

	b.StopTimer()
	client.FlushDB(ctx)
	client.Close()
}
