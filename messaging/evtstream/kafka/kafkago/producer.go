package kafka

import (
	"context"
	"errors"
	"fmt"
	"go-lab/messaging/evtstream"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers      []string
	KafkaVersion string
	MaxRetry     int

	AllowAutoTopicCreation bool
	EnableDebug            bool
}

type kafkaGoProducer struct {
	conf     *Config
	producer *kafkago.Writer
}

func NewProducer(conf *Config) (evtstream.Producer, error) {
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(conf.Brokers...),
		MaxAttempts:  conf.MaxRetry,
		RequiredAcks: kafkago.RequireAll,

		Balancer:               &kafkago.LeastBytes{},
		AllowAutoTopicCreation: conf.AllowAutoTopicCreation,
	}

	if conf.EnableDebug {
		w.Logger = kafkago.LoggerFunc(logf)
		w.ErrorLogger = kafkago.LoggerFunc(logf)
	}

	return &kafkaGoProducer{
		conf:     conf,
		producer: w,
	}, nil
}

func (p *kafkaGoProducer) Publish(ctx context.Context, msg *evtstream.Message) error {
	m := kafkago.Message{
		Topic: msg.Topic,
		Key:   []byte(msg.Key),
		Value: msg.Value,
	}

	var err error
	retries := p.conf.MaxRetry

	for i := range retries {
		attemptCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

		err := p.producer.WriteMessages(attemptCtx, m)
		cancel()

		if err == nil {
			return nil
		}

		if p.allowRetry(err) {
			log.Printf("kafka publish retry %d/%d: %s", i+1, retries, err.Error())
			time.Sleep(250 * time.Millisecond)
			continue
		}

		return err
	}

	return err
}

func (p *kafkaGoProducer) Close() error {
	return p.producer.Close()
}

func (p *kafkaGoProducer) allowRetry(err error) bool {
	return errors.Is(err, kafkago.LeaderNotAvailable) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, kafkago.UnknownTopicOrPartition)
}

func logf(msg string, a ...any) {
	fmt.Printf(msg, a...)
	fmt.Println()
}
