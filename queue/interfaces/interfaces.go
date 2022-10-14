package interfaces

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

//go:generate mockgen -destination=../../mocks/server.go -package=mocks . IServer,AmqpChannel,AmqpConnection,Connector,Publisher

type AmqpChannel interface {
	Cancel(consumer string, noWait bool) error
	NotifyClose(c chan *amqp.Error) chan *amqp.Error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

type AmqpDialer func(url string) (AmqpConnection, error)

type AmqpConnection interface {
	Close() error
	Channel() (AmqpChannel, error)
}

type Connector interface {
	Connect(url string) (AmqpConnection, error)
}

type Publisher interface {
	PublishWithContext(ctx context.Context, channel AmqpChannel, exchange, key string, msg amqp.Publishing) error
}

type IServer interface {
	Connector
	Publisher

	Stop() error
}
