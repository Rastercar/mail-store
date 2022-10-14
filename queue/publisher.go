package queue

import (
	"context"
	"mail-store-ms/queue/interfaces"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct{}

func (p *Publisher) PublishWithContext(ctx context.Context, channel interfaces.AmqpChannel, exchange, key string, msg amqp.Publishing) error {
	return channel.PublishWithContext(
		ctx,      // context
		exchange, // exchange
		key,      // key
		false,    // mandatory
		false,    // immediate
		msg,      // msg
	)
}
