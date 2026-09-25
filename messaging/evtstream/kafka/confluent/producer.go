package kafka

import (
	"context"
	"errors"
	"fmt"
	"go-lab/messaging/evtstream"
	"strings"

	confluent "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Config struct {
	Brokers      []string
	KafkaVersion string
	MaxRetry     int

	AllowAutoTopicCreation bool
	EnableDebug            bool
}

type confluentKafka struct {
	producer *confluent.Producer
}

func NewProducer(cfg *Config) (evtstream.Producer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers is empty")
	}

	config := &confluent.ConfigMap{
		"bootstrap.servers":        strings.Join(cfg.Brokers, ","),
		"retries":                  cfg.MaxRetry,
		"allow.auto.create.topics": cfg.AllowAutoTopicCreation,
	}

	if cfg.EnableDebug {
		_ = config.SetKey("debug", "broker,topic,msg")
	}

	p, err := confluent.NewProducer(config)
	if err != nil {
		return nil, err
	}

	return &confluentKafka{
		producer: p,
	}, nil
}

func (p *confluentKafka) Publish(ctx context.Context, msg *evtstream.Message) error {
	go func() {
		for e := range p.producer.Events() {
			switch ev := e.(type) {
			case *confluent.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	p.producer.Produce(&confluent.Message{
		TopicPartition: confluent.TopicPartition{Topic: &msg.Topic},
		Key:            []byte(msg.Key),
		Value:          []byte(msg.Value),
	}, nil)

	return nil
}

func (p *confluentKafka) Close() error {
	// Wait for message deliveries before shutting down
	p.producer.Flush(15 * 1000)
	p.producer.Close()

	return nil
}
