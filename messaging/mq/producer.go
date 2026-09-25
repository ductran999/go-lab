package mq

import "context"

type Producer interface {
	Publish(ctx context.Context, msg Message) error
	Close() error
}

type Message struct {
	Queue   string
	Key     string // optional (future exchange routing)
	Value   []byte
	Headers map[string]any // optional
}
