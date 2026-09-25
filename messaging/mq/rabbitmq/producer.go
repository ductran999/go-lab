package rabbitmq

import (
	"context"
	"go-lab/messaging/mq"
	"maps"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewProducer(url string) (*Producer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Producer{
		conn:    conn,
		channel: ch,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, msg mq.Message) error {
	// Declare queue (idempotent)
	_, err := p.channel.QueueDeclare(
		msg.Queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	headers := amqp.Table{}
	maps.Copy(headers, msg.Headers)

	return p.channel.PublishWithContext(
		ctx,
		"",        // default exchange
		msg.Queue, // routing key = queue
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/octet-stream",
			Headers:      headers,
			Body:         msg.Value,
		},
	)
}

func (p *Producer) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	return p.conn.Close()
}
