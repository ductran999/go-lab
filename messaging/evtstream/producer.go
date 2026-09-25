package evtstream

import "context"

type Message struct {
	Topic string
	Key   string
	Value []byte
}

type Producer interface {
	Publish(ctx context.Context, msg *Message) (err error)

	Close() error
}
