package lock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Lock 分布式锁
type Lock struct {
	client *redis.Client
	mu     sync.Mutex
}

// New 创建分布式锁实例
func New(client *redis.Client) *Lock {
	return &Lock{
		client: client,
	}
}

// Lock 获取锁（阻塞重试，直到成功或超时）
// key: 锁的 key
// expiration: 锁超时时间（0 则默认 30s）
// timeout: 获取锁的超时时间（0 则一直重试）
func (l *Lock) Lock(ctx context.Context, key string, expiration, timeout time.Duration) (token string, err error) {
	if expiration <= 0 {
		expiration = 30 * time.Second
	}

	token = uuid.New().String()

	var deadline time.Time
	var hasDeadline bool
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
		hasDeadline = true
	}

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		ok, err := l.client.SetNX(ctx, key, token, expiration).Result()
		if err != nil {
			return "", fmt.Errorf("获取锁失败: %w", err)
		}
		if ok {
			return token, nil
		}

		if hasDeadline && time.Now().After(deadline) {
			return "", fmt.Errorf("获取锁超时")
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// TryLock 尝试获取锁（不阻塞，立即返回）
// key: 锁的 key
// expiration: 锁超时时间（0 则默认 30s）
// 返回: token（成功时）、是否成功、错误
func (l *Lock) TryLock(ctx context.Context, key string, expiration time.Duration) (token string, ok bool, err error) {
	if expiration <= 0 {
		expiration = 30 * time.Second
	}

	token = uuid.New().String()

	ok, err = l.client.SetNX(ctx, key, token, expiration).Result()
	if err != nil {
		return "", false, fmt.Errorf("尝试获取锁失败: %w", err)
	}
	if !ok {
		return "", false, nil
	}

	return token, true, nil
}

// Unlock 释放锁（Lua 脚本保证原子性，只释放自己的锁）
// key: 锁的 key
// token: 加锁时返回的 token
func (l *Lock) Unlock(ctx context.Context, key, token string) error {
	// Lua 脚本：只有 value 匹配时才删除，防止误删别人的锁
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{key}, token).Result()
	if err != nil {
		return fmt.Errorf("释放锁失败: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("释放锁失败: token 不匹配或锁已过期")
	}

	return nil
}

// SpinLock 自旋锁（带超时，可指定重试间隔）
// key: 锁的 key
// expiration: 锁超时时间（0 则默认 30s）
// timeout: 获取锁的超时时间
// retryDelay: 重试间隔（0 则默认 100ms）
func (l *Lock) SpinLock(ctx context.Context, key string, expiration, timeout, retryDelay time.Duration) (token string, err error) {
	if expiration <= 0 {
		expiration = 30 * time.Second
	}
	if retryDelay <= 0 {
		retryDelay = 100 * time.Millisecond
	}

	token = uuid.New().String()

	deadline := time.Now().Add(timeout)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		ok, err := l.client.SetNX(ctx, key, token, expiration).Result()
		if err != nil {
			return "", fmt.Errorf("自旋锁失败: %w", err)
		}
		if ok {
			return token, nil
		}

		if time.Now().After(deadline) {
			return "", fmt.Errorf("自旋锁超时")
		}

		time.Sleep(retryDelay)
	}
}
