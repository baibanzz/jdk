package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRedis[T any] struct {
	client      *redis.Client
	ctx         context.Context
	defaultTime time.Duration
	cache       *CachePro[T]
}

// NewCacheRedis 创建一个新的Redis缓存实例
func NewCacheRedis[T any](client *redis.Client, defaultTimes, clearTime time.Duration) *CacheRedis[T] {
	return &CacheRedis[T]{
		client:      client,
		ctx:         context.Background(),
		defaultTime: defaultTimes,
		cache:       NewPro[T](defaultTimes, clearTime, nil),
	}
}

// Set 向本地缓存和Redis添加一个项目，替换任何现有项目
// 如果持续时间为0 (DefaultExpiration)，则使用默认过期时间（永不过期）
// 如果为-1 (NoExpiration)，则项目永不过期
// 如果大于0，则设置相应的TTL
func (c *CacheRedis[T]) Set(k string, x T, d time.Duration) error {
	// 写入Redis
	data, err := json.Marshal(x)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if d == DefaultExpiration {
		// 使用默认值，Redis中为0表示永不过期
		d = 0
	}

	var ttl time.Duration
	if d > 0 {
		ttl = d
	} else if d == NoExpiration {
		ttl = 0 // 永不过期
	} else {
		ttl = 0 // d为0或DefaultExpiration，永不过期
	}

	if ttl > 0 {
		return c.client.Set(c.ctx, k, data, ttl).Err()
	}
	if err := c.client.Set(c.ctx, k, data, 0).Err(); err != nil {
		return err
	}
	c.cache.Set(k, x, d)
	return nil
}

// SetDefault 向Redis缓存添加一个项目，使用默认过期时间（永不过期）
func (c *CacheRedis[T]) SetDefault(k string, x T) error {
	return c.Set(k, x, c.defaultTime)
}

// Add 仅当给定键不存在时，向Redis缓存添加项目
// 否则返回错误
func (c *CacheRedis[T]) Add(k string, x T, d time.Duration) error {
	data, err := json.Marshal(x)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if d == DefaultExpiration {
		d = 0
	}

	var ttl time.Duration
	if d > 0 {
		ttl = d
	} else if d == NoExpiration {
		ttl = 0
	} else {
		ttl = 0
	}

	// 使用SETNX命令（SET if Not eXists）
	if ttl > 0 {
		// Redis v6.2.0+ 支持SET命令的NX和EX选项一起使用
		// 为了兼容性，我们使用SET NX EX
		success, err := c.client.SetNX(c.ctx, k, data, ttl).Result()
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("item %s already exists", k)
		}
		return nil
	} else {
		success, err := c.client.SetNX(c.ctx, k, data, 0).Result()
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("item %s already exists", k)
		}
		return nil
	}
}

// Replace 仅当Redis缓存键已存在时，设置新值
// 否则返回错误
func (c *CacheRedis[T]) Replace(k string, x T, d time.Duration) error {
	// 首先检查键是否存在
	exists, err := c.client.Exists(c.ctx, k).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("item %s doesn't exist", k)
	}

	// 键存在，更新值
	return c.Set(k, x, d)
}

// Get 从本地缓存获取项目，如果本地缓存没有则从Redis获取
// 返回项目或零值，以及一个布尔值指示是否找到键
func (c *CacheRedis[T]) Get(k string) (T, bool) {
	// 先从本地缓存获取
	if value, found := c.cache.Get(k); found {
		return value, true
	}

	// 本地缓存没有，从Redis获取
	var zero T
	data, err := c.client.Get(c.ctx, k).Result()
	if err != nil {
		if err == redis.Nil {
			return zero, false
		}
		return zero, false
	}
	var value T
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return zero, false
	}

	// 将从Redis获取的值存入本地缓存
	c.cache.Set(k, value, c.defaultTime)
	return value, true
}

// GetWithExpiration 从Redis缓存返回项目及其过期时间
// 返回项目或零值，过期时间，以及一个布尔值指示是否找到键
func (c *CacheRedis[T]) GetWithExpiration(k string) (T, time.Time, bool) {
	var zero T
	// Redis的TTL命令返回剩余的生存时间（秒）
	ttl, err := c.client.TTL(c.ctx, k).Result()
	if err != nil {
		return zero, time.Time{}, false
	}

	// 检查键是否存在
	exists, err := c.client.Exists(c.ctx, k).Result()
	if err != nil || exists == 0 {
		return zero, time.Time{}, false
	}

	// 获取值
	data, err := c.client.Get(c.ctx, k).Result()
	if err != nil {
		return zero, time.Time{}, false
	}

	var value T
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return zero, time.Time{}, false
	}

	// 计算过期时间
	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	} else if ttl == -2 { // 键不存在
		return zero, time.Time{}, false
	} else if ttl == -1 { // 键存在但没有设置过期时间
		expiration = time.Time{}
	}

	return value, expiration, true
}

// Delete 从本地缓存和Redis删除项目
func (c *CacheRedis[T]) Delete(k string) error {
	// 删除Redis
	if err := c.client.Del(c.ctx, k).Err(); err != nil {
		return err
	}
	// 同时删除本地缓存
	c.cache.Delete(k)
	return nil
}

// DeleteExpired Redis会自动删除过期项目，此方法用于兼容API
// 实际上Redis会在访问时检查过期，并定期删除过期键
func (c *CacheRedis[T]) DeleteExpired() error {
	// Redis会自动处理过期键，这里不需要做任何事情
	// 但为了兼容API，我们可以返回nil
	return nil
}

// Items 获取所有未过期的键值对（注意：对于大型Redis数据库可能效率不高）
func (c *CacheRedis[T]) Items() (map[string]T, error) {
	// 获取所有键（在生产环境中应该使用SCAN而不是KEYS）
	keys, err := c.client.Keys(c.ctx, "*").Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]T)
	for _, key := range keys {
		// 检查键是否已过期（Redis会在获取时自动过滤）
		value, found := c.Get(key)
		if found {
			result[key] = value
		}
	}
	return result, nil
}

// ItemCount 返回Redis缓存中的项目数
func (c *CacheRedis[T]) ItemCount() (int64, error) {
	return c.client.DBSize(c.ctx).Result()
}

// Flush 清空本地缓存和当前数据库的所有键
func (c *CacheRedis[T]) Flush() error {
	// 清空Redis
	if err := c.client.FlushDB(c.ctx).Err(); err != nil {
		return err
	}
	// 清空本地缓存
	c.cache.Flush()
	return nil
}

// Increment 增加数值类型的值
func (c *CacheRedis[T]) Increment(k string, n int64) (int64, error) {
	return c.client.IncrBy(c.ctx, k, n).Result()
}

// IncrementFloat 增加浮点数类型的值
func (c *CacheRedis[T]) IncrementFloat(k string, n float64) (float64, error) {
	return c.client.IncrByFloat(c.ctx, k, n).Result()
}

// Decrement 减少数值类型的值
func (c *CacheRedis[T]) Decrement(k string, n int64) (int64, error) {
	return c.client.DecrBy(c.ctx, k, n).Result()
}

// OnEvicted 设置驱逐回调函数（Redis缓存不支持此功能，仅用于兼容API）
func (c *CacheRedis[T]) OnEvicted(f func(string, interface{})) {
	// Redis不支持驱逐回调，这里不实现
}

// Save 和 Load 方法不适用于Redis缓存，因为Redis已经是持久化的
// 这些方法仅用于兼容API
func (c *CacheRedis[T]) Save(w io.Writer) error {
	return fmt.Errorf("Save not supported for Redis cache")
}

func (c *CacheRedis[T]) SaveFile(fname string) error {
	return fmt.Errorf("SaveFile not supported for Redis cache")
}

func (c *CacheRedis[T]) Load(r io.Reader) error {
	return fmt.Errorf("Load not supported for Redis cache")
}

func (c *CacheRedis[T]) LoadFile(fname string) error {
	return fmt.Errorf("LoadFile not supported for Redis cache")
}
