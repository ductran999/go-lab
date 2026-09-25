package kafka

import (
	"context"
	"go-lab/messaging/evtstream"
	kafka "go-lab/messaging/evtstream/kafka/confluent"
	"log"
)

func RunConfluentKafkaGoExample() {
	config := &kafka.Config{
		Brokers:                []string{"localhost:29092"},
		KafkaVersion:           "4.0.0.0",
		MaxRetry:               3,
		AllowAutoTopicCreation: true,
		EnableDebug:            true,
	}

	topic := "confluent-kafka-go-topic"

	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalln(err)
	}
	defer producer.Close()

	msg := &evtstream.Message{
		Topic: topic,
		Key:   "user",
		Value: []byte("hello kafka from confluent"),
	}

	if err := producer.Publish(context.Background(), msg); err != nil {
		log.Println("failed to send message:" + err.Error())
	} else {
		log.Println("message sent to topic:" + topic)
	}
}
