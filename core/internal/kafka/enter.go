package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/baibanzz/jdk/model"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// Kafka 封装了 kafka-go 的客户端
type Kafka struct {
	cfg      model.Kafka
	dialer   *kafka.Dialer
	mu       sync.RWMutex
	closed   bool
}

// New 创建 Kafka 客户端
func New(cfg model.Kafka) (*Kafka, error) {
	if len(cfg.Addrs) == 0 {
		return nil, fmt.Errorf("Kafka 地址列表不能为空")
	}

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	// SASL 认证
	if cfg.Username != "" && cfg.Password != "" {
		dialer.SASLMechanism = plain.Mechanism{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	// TLS 配置
	if cfg.TLS.Enable {
		tlsConfig, err := buildTLSConfig(cfg.TLS)
		if err != nil {
			return nil, fmt.Errorf("构建 TLS 配置失败: %w", err)
		}
		dialer.TLS = tlsConfig
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", cfg.Addrs[0])
	if err != nil {
		return nil, fmt.Errorf("连接 Kafka 失败(%s): %w", cfg.Addrs[0], err)
	}
	_ = conn.Close()

	return &Kafka{
		cfg:    cfg,
		dialer: dialer,
	}, nil
}

// buildTLSConfig 构建 TLS 配置
func buildTLSConfig(cfg model.TLS) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipVerify,
	}

	// 加载 CA 证书
	if cfg.CAFile != "" {
		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("读取 CA 证书失败: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("解析 CA 证书失败")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// 加载客户端证书
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("加载客户端证书失败: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// getWriter 创建 writer（内部使用，不暴露）
func (k *Kafka) getWriter(topic string) *kafka.Writer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(k.cfg.Addrs...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		BatchBytes:   1048576, // 1MB
		Async:        false,
	}

	if k.dialer.SASLMechanism != nil || k.dialer.TLS != nil {
		w.Transport = &kafka.Transport{
			SASL: k.dialer.SASLMechanism,
			TLS:  k.dialer.TLS,
		}
	}

	return w
}

// SendMessage 发送单条消息
func (k *Kafka) SendMessage(topic string, key string, value []byte) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.closed {
		return fmt.Errorf("Kafka 客户端已关闭")
	}

	w := k.getWriter(topic)
	defer w.Close()

	msg := kafka.Message{
		Key:   []byte(key),
		Value: value,
	}

	return w.WriteMessages(context.Background(), msg)
}

// SendMessages 批量发送消息
func (k *Kafka) SendMessages(topic string, msgs ...kafka.Message) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.closed {
		return fmt.Errorf("Kafka 客户端已关闭")
	}

	w := k.getWriter(topic)
	defer w.Close()

	return w.WriteMessages(context.Background(), msgs...)
}

// Consumer 消费消息（自动提交偏移量）
// handler 返回 error 表示处理失败，不会提交偏移量
func (k *Kafka) Consumer(ctx context.Context, topic, group string, handler func(msg kafka.Message) error) error {
	k.mu.RLock()
	if k.closed {
		k.mu.RUnlock()
		return fmt.Errorf("Kafka 客户端已关闭")
	}
	k.mu.RUnlock()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     k.cfg.Addrs,
		Topic:       topic,
		GroupID:     group,
		MinBytes:    10,              // 10B
		MaxBytes:    10e6,            // 10MB
		MaxWait:     1 * time.Second, // 最大等待时间
		CommitInterval: 0,            // 同步提交
		StartOffset: kafka.LastOffset,
	})

	// SASL / TLS
	if k.dialer.SASLMechanism != nil || k.dialer.TLS != nil {
		r.SetOffset(kafka.LastOffset)
	}

	defer r.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := r.FetchMessage(ctx)
		if err != nil {
			return nil
		}

		if err := handler(msg); err != nil {
			return fmt.Errorf("处理消息失败: %w", err)
		}

		if err := r.CommitMessages(context.Background(), msg); err != nil {
			return fmt.Errorf("提交偏移量失败: %w", err)
		}
	}
}

// ConsumerManual 消费消息（手动提交偏移量）
// handler 返回 error 表示处理失败，不会提交偏移量
// 与 Consumer 的区别：handler 返回 error 后不会终止消费，而是继续下一条
func (k *Kafka) ConsumerManual(ctx context.Context, topic, group string, handler func(msg kafka.Message) error) error {
	k.mu.RLock()
	if k.closed {
		k.mu.RUnlock()
		return fmt.Errorf("Kafka 客户端已关闭")
	}
	k.mu.RUnlock()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     k.cfg.Addrs,
		Topic:       topic,
		GroupID:     group,
		MinBytes:    10,
		MaxBytes:    10e6,
		MaxWait:     1 * time.Second,
		CommitInterval: 0,
		StartOffset: kafka.LastOffset,
	})

	defer r.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := r.FetchMessage(ctx)
		if err != nil {
			return nil
		}

		// 即使 handler 失败也不终止，继续消费
		if err := handler(msg); err != nil {
			continue
		}

		if err := r.CommitMessages(context.Background(), msg); err != nil {
			return fmt.Errorf("提交偏移量失败: %w", err)
		}
	}
}

// CreateTopic 创建主题
func (k *Kafka) CreateTopic(topic string, partitions, replicationFactor int) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.closed {
		return fmt.Errorf("Kafka 客户端已关闭")
	}

	conn, err := k.dialer.Dial("tcp", k.cfg.Addrs[0])
	if err != nil {
		return fmt.Errorf("连接 Kafka 失败: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("获取 Controller 失败: %w", err)
	}

	controllerConn, err := k.dialer.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("连接 Controller 失败: %w", err)
	}
	defer controllerConn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	return controllerConn.CreateTopics(topicConfig)
}

// DeleteTopic 删除主题
func (k *Kafka) DeleteTopic(topic string) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.closed {
		return fmt.Errorf("Kafka 客户端已关闭")
	}

	conn, err := k.dialer.Dial("tcp", k.cfg.Addrs[0])
	if err != nil {
		return fmt.Errorf("连接 Kafka 失败: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("获取 Controller 失败: %w", err)
	}

	controllerConn, err := k.dialer.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("连接 Controller 失败: %w", err)
	}
	defer controllerConn.Close()

	return controllerConn.DeleteTopics(topic)
}

// Close 关闭 Kafka 客户端
func (k *Kafka) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.closed {
		return nil
	}
	k.closed = true

	return nil
}
