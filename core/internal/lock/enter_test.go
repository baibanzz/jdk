package lock

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisAddr = envOrDefault("REDIS_ADDR", "localhost:6379")
	redisPass = envOrDefault("REDIS_PASS", "88888888")
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       1, // 使用 DB 1 避免影响其他数据
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("连接 Redis 失败: %v", err)
	}

	return client
}

func TestNewLock(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()

	l := New(client)
	if l == nil {
		t.Fatal("创建 Lock 失败")
	}
	t.Log("Lock 创建成功")
}

func TestLock_TryLockAndUnlock(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()

	l := New(client)
	ctx := context.Background()
	key := "test:lock:try"

	// 尝试获取锁
	token, ok, err := l.TryLock(ctx, key, 10*time.Second)
	if err != nil {
		t.Fatalf("TryLock 失败: %v", err)
	}
	if !ok {
		t.Fatal("TryLock 应返回成功")
	}
	t.Logf("获取锁成功, token=%s", token)

	// 再次尝试获取同一把锁应失败
	_, ok, err = l.TryLock(ctx, key, 10*time.Second)
	if err != nil {
		t.Fatalf("TryLock 失败: %v", err)
	}
	if ok {
		t.Fatal("重复获取锁应返回失败")
	}
	t.Log("重复获取锁返回预期失败")

	// 释放锁
	err = l.Unlock(ctx, key, token)
	if err != nil {
		t.Fatalf("Unlock 失败: %v", err)
	}
	t.Log("释放锁成功")

	// 释放后应能再次获取
	_, ok, err = l.TryLock(ctx, key, 10*time.Second)
	if err != nil {
		t.Fatalf("TryLock 失败: %v", err)
	}
	if !ok {
		t.Fatal("释放后应能再次获取锁")
	}
	t.Log("释放后再次获取锁成功")
}

func TestLock_UnlockWithWrongToken(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()

	l := New(client)
	ctx := context.Background()
	key := "test:lock:wrong-token"

	// 获取锁
	token, ok, err := l.TryLock(ctx, key, 10*time.Second)
	if err != nil {
		t.Fatalf("TryLock 失败: %v", err)
	}
	if !ok {
		t.Fatal("TryLock 应返回成功")
	}

	// 用错误的 token 释放锁应失败
	err = l.Unlock(ctx, key, "wrong-token")
	if err == nil {
		t.Fatal("使用错误 token 释放锁应返回错误")
	}
	t.Logf("错误 token 释放锁返回预期错误: %v", err)

	// 用正确的 token 释放锁
	err = l.Unlock(ctx, key, token)
	if err != nil {
		t.Fatalf("Unlock 失败: %v", err)
	}
	t.Log("正确 token 释放锁成功")
}

func TestLock_Lock(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()

	l := New(client)
	ctx := context.Background()
	key := "test:lock:blocking"

	// 获取锁
	token, err := l.Lock(ctx, key, 10*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("Lock 失败: %v", err)
	}
	t.Logf("获取锁成功, token=%s", token)

	// 释放锁
	err = l.Unlock(ctx, key, token)
	if err != nil {
		t.Fatalf("Unlock 失败: %v", err)
	}
	t.Log("释放锁成功")
}

func TestLock_ConcurrentLock(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()

	l := New(client)
	ctx := context.Background()
	key := "test:lock:concurrent"

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			token, ok, err := l.TryLock(ctx, key, 5*time.Second)
			if err != nil {
				t.Logf("goroutine %d TryLock 错误: %v", id, err)
				return
			}
			if ok {
				mu.Lock()
				successCount++
				mu.Unlock()
				t.Logf("goroutine %d 获取锁成功", id)
				// 模拟业务处理
				time.Sleep(100 * time.Millisecond)
				_ = l.Unlock(ctx, key, token)
			} else {
				t.Logf("goroutine %d 获取锁失败（被其他 goroutine 持有）", id)
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Fatalf("只有一个 goroutine 应获取到锁, 实际=%d", successCount)
	}
	t.Logf("并发锁测试通过, 只有 %d 个 goroutine 获取到锁", successCount)
}

func TestLock_SpinLock(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()

	l := New(client)
	ctx := context.Background()
	key := "test:lock:spin"

	// 先获取锁占住
	token1, err := l.Lock(ctx, key, 3*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("Lock 失败: %v", err)
	}
	t.Log("第一个 goroutine 获取锁成功")

	// 另一个 goroutine 自旋等待
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		token2, err := l.SpinLock(ctx, key, 10*time.Second, 10*time.Second, 100*time.Millisecond)
		if err != nil {
			t.Errorf("SpinLock 失败: %v", err)
			return
		}
		t.Logf("第二个 goroutine 自旋获取锁成功, token=%s", token2)
		_ = l.Unlock(ctx, key, token2)
	}()

	// 等待 1 秒后释放锁
	time.Sleep(1 * time.Second)
	_ = l.Unlock(ctx, key, token1)
	t.Log("第一个 goroutine 释放锁")

	wg.Wait()
	t.Log("自旋锁测试通过")
}

func TestLock_LockTimeout(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()

	l := New(client)
	ctx := context.Background()
	key := "test:lock:timeout"

	// 先获取锁占住
	token1, err := l.Lock(ctx, key, 10*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("Lock 失败: %v", err)
	}
	t.Log("第一个 goroutine 获取锁成功")

	// 尝试获取锁，超时 1 秒
	_, err = l.Lock(ctx, key, 10*time.Second, 1*time.Second)
	if err == nil {
		t.Fatal("锁被占用时应超时")
	}
	t.Logf("锁超时返回预期错误: %v", err)

	// 清理
	_ = l.Unlock(ctx, key, token1)
}
