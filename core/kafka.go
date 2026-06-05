package core

import (
	"github.com/baibanzz/jdk/core/internal/kafka"
	"github.com/baibanzz/jdk/model"
)

type Kafka = kafka.Kafka

func NewKafka(config model.Kafka) (*Kafka, error) {
	if len(config.Addrs) == 0 {
		config.Addrs = []string{"localhost:9092"}
	}
	return kafka.New(config)
}
