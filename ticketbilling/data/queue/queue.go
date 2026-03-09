package queue

import "context"

type Consumer interface {
	// ConsumeAsync es el método que inicia la escucha de mensajes.
	// ACK/NACK.
	ConsumeAsync(handler func(delivery Delivery)) error
}

type Delivery interface {
	Body() []byte
	Ack() error
	Nack(requeue bool) error
}

type Producer interface {
	Publish(ctx context.Context, payload interface{}) error
	PublishWithRetry(ctx context.Context, payload interface{}, delaySeconds int) error
}

type RSGQueue interface {
	Consumer
	Producer
	Close()
}
