package queue

import (
	"context"
	"log"
	"mail-store-ms/config"
	"mail-store-ms/queue/interfaces"
	"mail-store-ms/tracer"

	amqp "github.com/rabbitmq/amqp091-go"
)

//go:generate mockgen -destination=../mocks/amqp.go -package=mocks github.com/rabbitmq/amqp091-go Acknowledger

type RpcRes struct {
	Error        error  // nil if the delivery was successfully processed and should be ackd or rejected otherwise
	ResponseBody []byte // nil or a json serializable struct to send on the amqp response body
}

type RpcRouter map[string]RpcCallHandler

type RpcCallHandler func(context.Context, *amqp.Delivery) RpcRes

type Server struct {
	interfaces.Connector
	interfaces.Publisher

	cfg            config.RmqConfig
	conn           interfaces.AmqpConnection
	channel        interfaces.AmqpChannel
	deliveries     <-chan amqp.Delivery
	notifyClose    chan *amqp.Error
	DeliveryRouter RpcRouter
}

func New(cfg config.RmqConfig) *Server {
	return &Server{
		cfg:       cfg,
		Connector: &Connector{},
		Publisher: &Publisher{},
	}
}

func (s *Server) Start() {
	go func() {
		for {
			s.connect()

			connectionError, chanClosed := <-s.notifyClose

			// connection error is nil and chanClosed is false when
			// the connection was closed manually with client code
			if connectionError != nil {
				log.Printf("[ RMQ ] connection error: %v \n", connectionError)
			}

			if !chanClosed {
				return
			}
		}
	}()
}

func (s *Server) Stop() error {
	log.Printf("[ RMQ ] closing connections")

	if s.conn != nil {
		return s.conn.Close()
	}

	return nil
}

// Publishes a publishing with the through the main channel using the Publisher interface
func (s *Server) Publish(ctx context.Context, exchange, key string, publishing amqp.Publishing) error {
	ctx, span := tracer.NewSpan(ctx, "queue", "Publish")
	defer span.End()

	err := s.Publisher.PublishWithContext(ctx, s.channel, exchange, key, publishing)

	if err != nil {
		tracer.AddSpanErrorAndFail(span, err, "publish failure")
	} else {
		tracer.RecordSpanSuccess(span, "publish successfull")
	}

	return err
}
