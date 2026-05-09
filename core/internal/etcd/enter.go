package etcd

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/baibanzz/jdk/model"
	"go.etcd.io/etcd/client/v3"
)

type Etcd struct {
	client   *clientv3.Client
	cfg      clientv3.Config
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	stopChan chan struct{}
	closed   bool
}

func New(d model.Etcd) (*Etcd, error) {
	cfg := clientv3.Config{
		Endpoints:   d.Host,
		Username:    d.User,
		Password:    d.Pass,
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("连接 Etcd 失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Etcd{
		client:   client,
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
		stopChan: make(chan struct{}),
	}, nil
}

// reConnect 重新连接 etcd
func (etcd *Etcd) reConnect() error {
	etcd.mu.Lock()
	defer etcd.mu.Unlock()

	// 关闭旧客户端
	if etcd.client != nil {
		etcd.client.Close()
	}

	// 创建新客户端
	client, err := clientv3.New(etcd.cfg)
	if err != nil {
		return fmt.Errorf("重连 Etcd 失败: %w", err)
	}

	etcd.client = client
	return nil
}

// checkConnection 检查并维护连接，断线自动重连
func (etcd *Etcd) checkConnection() error {
	etcd.mu.RLock()
	client := etcd.client
	etcd.mu.RUnlock()

	// 检查连接状态
	if client == nil {
		return etcd.reConnect()
	}

	// 尝试 ping 检测连接
	ctx, cancel := context.WithTimeout(etcd.ctx, 2*time.Second)
	defer cancel()

	_, err := client.Status(ctx, client.Endpoints()[0])
	if err != nil {
		log.Printf("Etcd 连接断开，尝试重连: %v", err)
		return etcd.reConnect()
	}

	return nil
}

// AutoPut 自动重复提交 ttl 10s，包含断线自动重连和每5秒自动发送
func (etcd *Etcd) AutoPut(K string, V []byte) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-etcd.ctx.Done():
				return
			case <-etcd.stopChan:
				return
			case <-ticker.C:
				// 检查并维护连接
				if err := etcd.checkConnection(); err != nil {
					log.Printf("Etcd AutoPut 连接检查失败: %v", err)
					continue
				}

				// 创建 10 秒 TTL 的 lease
				etcd.mu.RLock()
				client := etcd.client
				etcd.mu.RUnlock()

				if client == nil {
					continue
				}

				ctx, cancel := context.WithTimeout(etcd.ctx, 3*time.Second)
				resp, err := client.Grant(ctx, 10)
				if err != nil {
					cancel()
					log.Printf("Etcd AutoPut 创建 Lease 失败: %v", err)
					continue
				}

				// 使用 lease 提交 key-value
				_, err = client.Put(ctx, K, string(V), clientv3.WithLease(resp.ID))
				cancel()
				if err != nil {
					log.Printf("Etcd AutoPut 提交失败: %v", err)
				}
			}
		}
	}()
}

// Put 提交，带 TTL
func (etcd *Etcd) Put(K string, V []byte, ttl time.Duration) error {
	if ttl > 0 {
		// 创建指定 TTL 的 lease
		grantResp, err := etcd.client.Grant(etcd.ctx, int64(ttl.Seconds()))
		if err != nil {
			return err
		}
		_, err = etcd.client.Put(etcd.ctx, K, string(V), clientv3.WithLease(grantResp.ID))
		return err
	}
	// 无 TTL 直接提交
	_, err := etcd.client.Put(etcd.ctx, K, string(V))
	return err
}

func (etcd *Etcd) Delete(K string) error {
	_, err := etcd.client.Delete(etcd.ctx, K)
	return err
}

func (etcd *Etcd) GetList(K string) (map[string][]byte, error) {
	resp, err := etcd.client.Get(etcd.ctx, K, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = kv.Value
	}
	return result, nil
}

func (etcd *Etcd) Get(K string) ([]byte, error) {
	resp, err := etcd.client.Get(etcd.ctx, K)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	return resp.Kvs[0].Value, nil
}

func (etcd *Etcd) Close() error {
	etcd.mu.Lock()
	defer etcd.mu.Unlock()

	if etcd.closed {
		return nil
	}
	etcd.closed = true

	// 停止所有 goroutine
	etcd.cancel()
	close(etcd.stopChan)

	return etcd.client.Close()
}
