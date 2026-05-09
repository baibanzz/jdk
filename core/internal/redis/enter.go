package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/baibanzz/jdk/model"
	"github.com/redis/go-redis/v9"
)

func NewRedis(r model.Redis) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     r.Addr(),
		Password: r.Password,
		DB:       r.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return client, nil
}
