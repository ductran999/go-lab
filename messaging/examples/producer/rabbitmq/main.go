package main

import (
	"context"
	"go-lab/messaging/mq"
	"go-lab/messaging/mq/rabbitmq"
	"log"
)

func main() {
	producer, err := rabbitmq.NewProducer(
		"amqp://guest:guest@localhost:5672/",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	err = producer.Publish(context.Background(), mq.Message{
		Queue: "jobs.email",
		Value: []byte(`{"to":"user@example.com"}`),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Publish message done!")
}
