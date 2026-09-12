package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CacheRedisDB[T any] struct {
	rs   *redis.Client
	db   *gorm.DB
	back func(db *gorm.DB, key ...any) (*T, error)
	key  string
	ctx  context.Context
	ttl  time.Duration
}

type CacheRedisDBOpt[T any] struct {
	Rs   *redis.Client                             //redis客户端
	Key  string                                    // xxx:xxx:%s:%s... 存储的键
	Back func(db *gorm.DB, key ...any) (*T, error) //失败后的回调用于使用mysql之类的进行兜底 如果返回err为nil 则会写入redis
	Db   *gorm.DB                                  //使用back必填的数据库连接
	TTL  time.Duration                             //默认时间
}

// NewCacheRedisDB 创建redis缓存
func NewCacheRedisDB[T any](opt CacheRedisDBOpt[T]) *CacheRedisDB[T] {
	return &CacheRedisDB[T]{
		rs:   opt.Rs,
		key:  opt.Key,
		back: opt.Back,
		db:   opt.Db,
		ttl:  opt.TTL,
		ctx:  context.Background(),
	}
}

func (r *CacheRedisDB[T]) createKey(key []any) string {
	if len(key) == 0 {
		return r.key
	}
	return fmt.Sprintf(r.key, key...)
}

// err redis 未命中或出错时走 back 兜底并回写缓存; 命中时直接返回 nil 表示无需兜底
func (r *CacheRedisDB[T]) err(rerr error, key []any) (*T, error) {
	if rerr == nil {
		return nil, nil
	}
	if r.back == nil {
		return nil, rerr
	}
	back, berr := r.back(r.db, key...)
	if berr != nil {
		return nil, berr
	}
	if back == nil {
		return nil, rerr
	}
	if serr := r.Set(*back, key...); serr != nil {
		return nil, serr
	}
	return back, nil
}

func (r *CacheRedisDB[T]) Get(key ...any) (*T, error) {
	ckey := r.createKey(key)

	switch ret := any(new(T)).(type) {
	case *string:
		var err error
		*ret, err = r.rs.Get(r.ctx, ckey).Result()
		if d, err := r.err(err, key); err == nil && d != nil {
			return d, nil
		}
		if err != nil {
			return nil, err
		}
		return any(ret).(*T), nil
	default:
		bytes, err := r.rs.Get(r.ctx, ckey).Bytes()
		if d, err := r.err(err, key); err == nil && d != nil {
			return d, nil
		}
		if err != nil {
			return nil, err
		}
		if len(bytes) == 0 {
			return nil, nil
		}
		if err := json.Unmarshal(bytes, ret); err != nil {
			return nil, err
		}
		return ret.(*T), nil
	}
}

// SetTTL ttl填0为默认值 ttl -1为永久
func (r *CacheRedisDB[T]) SetTTL(val T, ttl time.Duration, key ...any) error {
	var t time.Duration
	if ttl == 0 {
		t = r.ttl
	} else if ttl > 0 {
		t = ttl
	} else {
		t = 0
	}
	switch any(val).(type) {
	case string:
		return r.rs.Set(r.ctx, r.createKey(key), val, t).Err()
	default:
		marshal, err := json.Marshal(val)
		if err != nil {
			return err
		}
		return r.rs.Set(r.ctx, r.createKey(key), marshal, t).Err()
	}
}

func (r *CacheRedisDB[T]) Set(val T, key ...any) error {
	return r.SetTTL(val, r.ttl, key...)
}

func (r *CacheRedisDB[T]) Del(key ...any) error {
	return r.rs.Del(r.ctx, r.createKey(key)).Err()
}
