package core

import (
	"github.com/baibanzz/jdk/core/internal/lock"
	"github.com/redis/go-redis/v9"
)

type Lock = lock.Lock

func NewLock(client *redis.Client) *Lock {
	return lock.New(client)
}
