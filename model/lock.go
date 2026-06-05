package model

import "time"

// Lock 分布式锁配置
//
// yaml 示例:
//
//	key: my-lock
//	expiration: 30s
//	retry: 3
//	retryDelay: 100ms
type Lock struct {
	// Key 锁的 key
	// yaml: key: my-lock
	Key string `json:"key" yaml:"key"`

	// Expiration 锁超时时间，默认 30s
	// yaml: expiration: 30s
	Expiration time.Duration `json:"expiration" yaml:"expiration"`

	// Retry 重试次数，默认 3
	// yaml: retry: 3
	Retry int `json:"retry" yaml:"retry"`

	// RetryDelay 重试间隔，默认 100ms
	// yaml: retryDelay: 100ms
	RetryDelay time.Duration `json:"retryDelay" yaml:"retryDelay"`
}
