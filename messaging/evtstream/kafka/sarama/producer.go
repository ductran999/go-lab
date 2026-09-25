package kafka

import (
	"context"
	"go-lab/messaging/evtstream"
	"log"
	"os"

	"github.com/IBM/sarama"
)

type Config struct {
	Brokers      []string
	KafkaVersion string
	MaxRetry     int

	AllowAutoTopicCreation bool
	EnableDebug            bool
}

type saramaProducer struct {
	producer sarama.SyncProducer
}

func NewProducer(conf *Config) (evtstream.Producer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V4_0_0_0

	config.Producer.Retry.Max = conf.MaxRetry
	config.Metadata.AllowAutoTopicCreation = conf.AllowAutoTopicCreation

	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	if conf.EnableDebug {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	producer, err := sarama.NewSyncProducer(conf.Brokers, config)
	if err != nil {
		return nil, err
	}

	return &saramaProducer{producer: producer}, nil
}

func (sp *saramaProducer) Publish(ctx context.Context, msg *evtstream.Message) error {
	m := &sarama.ProducerMessage{
		Topic: msg.Topic,
		Key:   sarama.StringEncoder(msg.Key),
		Value: sarama.StringEncoder(msg.Value),
	}

	_, _, err := sp.producer.SendMessage(m)
	if err != nil {
		return err
	}

	return nil
}

func (p *saramaProducer) Close() error {
	return p.producer.Close()
}
