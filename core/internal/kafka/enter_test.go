package kafka

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/baibanzz/jdk/model"
	"github.com/segmentio/kafka-go"
)

var (
	kafkaHost  = envOrDefault("KAFKA_HOST", "localhost:9092")
	testTopic  = envOrDefault("KAFKA_TEST_TOPIC", "test-topic")
	testGroup  = envOrDefault("KAFKA_TEST_GROUP", "test-group")
	testKey    = "test-key"
	testValue  = []byte("hello kafka")

	topicCounter int
)

func uniqueTopic() string {
	topicCounter++
	return fmt.Sprintf("%s-%d", testTopic, topicCounter)
}

func uniqueGroup() string {
	topicCounter++
	return fmt.Sprintf("%s-%d", testGroup, topicCounter)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func newTestKafka(t *testing.T) *Kafka {
	t.Helper()

	k, err := New(model.Kafka{
		Addrs: []string{kafkaHost},
	})
	if err != nil {
		t.Fatalf("创建 Kafka 客户端失败: %v", err)
	}
	return k
}

// createTopicAndWait 创建主题并等待就绪
func createTopicAndWait(t *testing.T, k *Kafka, topic string) {
	t.Helper()

	err := k.CreateTopic(topic, 1, 1)
	if err != nil {
		t.Fatalf("创建主题失败: %v", err)
	}

	// 等待 topic 就绪（最多等 5 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("等待主题 %s 就绪超时", topic)
		default:
		}

		// 通过 Dial 连接并获取 topic 的 partitions 来检测是否就绪
		conn, err := k.dialer.DialContext(ctx, "tcp", k.cfg.Addrs[0])
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		partitions, err := conn.ReadPartitions(topic)
		conn.Close()
		if err == nil && len(partitions) > 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestNewKafka(t *testing.T) {
	k, err := New(model.Kafka{
		Addrs: []string{kafkaHost},
	})
	if err != nil {
		t.Fatalf("创建 Kafka 客户端失败: %v", err)
	}
	defer k.Close()
	t.Logf("Kafka 客户端创建成功, 地址: %s", kafkaHost)
}

func TestKafka_SendMessage(t *testing.T) {
	k := newTestKafka(t)
	defer k.Close()

	topic := uniqueTopic()
	createTopicAndWait(t, k, topic)

	err := k.SendMessage(topic, testKey, testValue)
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}
	t.Logf("发送消息成功, topic=%s, key=%s, value=%s", topic, testKey, string(testValue))
}

func TestKafka_SendAndConsume(t *testing.T) {
	k := newTestKafka(t)
	defer k.Close()

	topic := uniqueTopic()
	group := uniqueGroup()
	createTopicAndWait(t, k, topic)

	// 先启动消费者 goroutine，再发送消息
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		wg     sync.WaitGroup
		gotMsg sync.WaitGroup
	)
	wg.Add(1)
	gotMsg.Add(1)

	go func() {
		defer wg.Done()
		err := k.Consumer(ctx, topic, group, func(msg kafka.Message) error {
			t.Logf("消费到消息: key=%s, value=%s, partition=%d, offset=%d",
				string(msg.Key), string(msg.Value), msg.Partition, msg.Offset)
			if string(msg.Key) != testKey {
				t.Errorf("key 不匹配, 期望=%s, 实际=%s", testKey, string(msg.Key))
			}
			if string(msg.Value) != string(testValue) {
				t.Errorf("value 不匹配, 期望=%s, 实际=%s", string(testValue), string(msg.Value))
			}
			gotMsg.Done()
			return nil
		})
		if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
			t.Errorf("消费消息失败: %v", err)
		}
	}()

	// 等待消费者启动
	time.Sleep(500 * time.Millisecond)

	// 发送消息
	err := k.SendMessage(topic, testKey, testValue)
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}
	t.Log("发送消息成功")

	// 等待消费到消息
	gotMsg.Wait()
	cancel()
	wg.Wait()
	t.Log("消费消息成功")
}

func TestKafka_SendMessages(t *testing.T) {
	k := newTestKafka(t)
	defer k.Close()

	topic := uniqueTopic()
	createTopicAndWait(t, k, topic)

	messages := []kafka.Message{
		{Key: []byte("k1"), Value: []byte("v1")},
		{Key: []byte("k2"), Value: []byte("v2")},
		{Key: []byte("k3"), Value: []byte("v3")},
	}

	err := k.SendMessages(topic, messages...)
	if err != nil {
		t.Fatalf("批量发送消息失败: %v", err)
	}
	t.Logf("批量发送 %d 条消息成功", len(messages))
}

func TestKafka_ConsumerManual(t *testing.T) {
	k := newTestKafka(t)
	defer k.Close()

	topic := uniqueTopic()
	group := uniqueGroup()
	createTopicAndWait(t, k, topic)

	// 先启动消费者 goroutine，再发送消息
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		wg     sync.WaitGroup
		gotMsg sync.WaitGroup
	)
	wg.Add(1)
	gotMsg.Add(1)

	go func() {
		defer wg.Done()
		err := k.ConsumerManual(ctx, topic, group, func(msg kafka.Message) error {
			t.Logf("手动提交消费到消息: key=%s, value=%s", string(msg.Key), string(msg.Value))
			gotMsg.Done()
			return nil
		})
		if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
			t.Errorf("手动消费消息失败: %v", err)
		}
	}()

	// 等待消费者启动
	time.Sleep(500 * time.Millisecond)

	// 发送消息
	err := k.SendMessage(topic, "manual-key", []byte("manual-value"))
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}
	t.Log("发送消息成功")

	// 等待消费到消息
	gotMsg.Wait()
	cancel()
	wg.Wait()
	t.Log("手动提交消费成功")
}

func TestKafka_CreateAndDeleteTopic(t *testing.T) {
	k := newTestKafka(t)
	defer k.Close()

	newTopic := "test-topic-" + time.Now().Format("150405")

	// 创建主题
	err := k.CreateTopic(newTopic, 3, 1)
	if err != nil {
		t.Fatalf("创建主题失败: %v", err)
	}
	t.Logf("创建主题成功: %s (partitions=3, replicationFactor=1)", newTopic)

	// 删除主题
	err = k.DeleteTopic(newTopic)
	if err != nil {
		t.Fatalf("删除主题失败: %v", err)
	}
	t.Logf("删除主题成功: %s", newTopic)
}

func TestKafka_Close(t *testing.T) {
	k := newTestKafka(t)

	// 关闭
	err := k.Close()
	if err != nil {
		t.Fatalf("关闭 Kafka 客户端失败: %v", err)
	}
	t.Log("Kafka 客户端关闭成功")

	// 关闭后再操作应返回错误
	err = k.SendMessage(testTopic, testKey, testValue)
	if err == nil {
		t.Fatal("关闭后发送消息应返回错误")
	}
	t.Logf("关闭后操作返回预期错误: %v", err)
}
