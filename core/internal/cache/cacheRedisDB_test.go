package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func newRedisDBTestClient(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{
		Addr:     addrhost,
		Password: password,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis服务器未运行，跳过测试")
	}
	client.FlushDB(context.Background())
	t.Cleanup(func() {
		client.FlushDB(context.Background())
		client.Close()
	})
	return client
}

// TestCacheRedisDBStringHit string 类型命中缓存必须返回真实值（回归：命中曾返回 nil）
func TestCacheRedisDBStringHit(t *testing.T) {
	client := newRedisDBTestClient(t)

	c := NewRedisCache[string](CacheRedisDBOpt[string]{
		Rs:  client,
		Key: "test:db:string",
		TTL: 5 * time.Minute,
	})

	if err := c.Set("hello"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, err := c.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil || *got != "hello" {
		t.Fatalf("Get returned wrong value: got %v, want hello", got)
	}
}

// TestCacheRedisDBStructHit 结构体命中走 JSON 路径
func TestCacheRedisDBStructHit(t *testing.T) {
	client := newRedisDBTestClient(t)

	type Person struct {
		Name string
		Age  int
	}
	c := NewRedisCache[Person](CacheRedisDBOpt[Person]{
		Rs:  client,
		Key: "test:db:person:%s",
		TTL: 5 * time.Minute,
	})

	want := Person{Name: "Alice", Age: 30}
	if err := c.Set(want, "1"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, err := c.Get("1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil || *got != want {
		t.Fatalf("Get returned wrong value: got %+v, want %+v", got, want)
	}
}

// TestCacheRedisDBMissBack 未命中走 back 兜底并回写，第二次 Get 不再调 back
func TestCacheRedisDBMissBack(t *testing.T) {
	client := newRedisDBTestClient(t)

	backCalls := 0
	c := NewRedisCache[string](CacheRedisDBOpt[string]{
		Rs:  client,
		Key: "test:db:back:%s",
		TTL: 5 * time.Minute,
		Back: func(db *gorm.DB, key ...any) (*string, error) {
			backCalls++
			if len(key) == 0 || key[0] != "k1" {
				t.Errorf("back key wrong: got %v, want [k1]", key)
			}
			s := "fromdb"
			return &s, nil
		},
	})

	got, err := c.Get("k1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil || *got != "fromdb" {
		t.Fatalf("Get returned wrong value: got %v, want fromdb", got)
	}
	if backCalls != 1 {
		t.Fatalf("back should be called once, got %d", backCalls)
	}

	// 回写后命中缓存，不再调 back
	got, err = c.Get("k1")
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}
	if got == nil || *got != "fromdb" {
		t.Fatalf("second Get wrong value: got %v", got)
	}
	if backCalls != 1 {
		t.Fatalf("back should still be called once, got %d", backCalls)
	}
}

// TestCacheRedisDBMissNoBack 无 back 时未命中返回 redis 错误
func TestCacheRedisDBMissNoBack(t *testing.T) {
	client := newRedisDBTestClient(t)

	c := NewRedisCache[string](CacheRedisDBOpt[string]{
		Rs:  client,
		Key: "test:db:noback",
		TTL: 5 * time.Minute,
	})

	_, err := c.Get()
	if err == nil {
		t.Fatal("Get on missing key without back should return error")
	}
}

// TestCacheRedisDBDel Del 后 Get 应未命中
func TestCacheRedisDBDel(t *testing.T) {
	client := newRedisDBTestClient(t)

	c := NewRedisCache[string](CacheRedisDBOpt[string]{
		Rs:  client,
		Key: "test:db:del",
		TTL: 5 * time.Minute,
	})

	if err := c.Set("v"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := c.Del(); err != nil {
		t.Fatalf("Del failed: %v", err)
	}
	if _, err := c.Get(); err == nil {
		t.Fatal("Get after Del should return error")
	}
}
