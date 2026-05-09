package core

import (
	"time"

	"github.com/baibanzz/jdk/core/internal/cache"
	"github.com/redis/go-redis/v9"
)

// NewCache 创建一个普通的Cache
func NewCache(defaultExpiration, cleanupInterval time.Duration) *cache.Cache {
	return cache.New(defaultExpiration, cleanupInterval)
}

// NewCachePro 创建一个加强版的Cache
func NewCachePro[T any](defaultExpiration, cleanupInterval time.Duration, DelFunc func(T)) *cache.CachePro[T] {
	return cache.NewPro(defaultExpiration, cleanupInterval, DelFunc)
}

// NewCacheRedis 创建一个新的Redis缓存实例
func NewCacheRedis[T any](client *redis.Client, defaultTimes, clearTime time.Duration) *cache.CacheRedis[T] {
	return cache.NewCacheRedis[T](client, defaultTimes, clearTime)
}
