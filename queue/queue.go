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

// // TODO: CHECK IF IM NEEDED
// // Publishes a publishing through the main channel using the Publisher interface to a queue using
// // the direct exchange and consumes the response queue until a response is recieved or timeouts
// //
// // Before sending, the publishing type, correlationId and replyTo fields are changed to the operation,
// // a new uuid and the server rpc response queue respectively
// func (s *Server) Rpc(ctx context.Context, queueName, operation string, publishing amqp.Publishing) (*amqp.Delivery, error) {
// 	ctx, span := tracer.NewSpan(ctx, "queue", "Rpc")
// 	defer span.End()

// 	correlationId := uuid.NewString()

// 	responses, err := s.channel.Consume(s.cfg.RpcResponseQueue, correlationId, false, false, false, false, nil)
// 	if err != nil {
// 		tracer.AddSpanErrorAndFail(span, err, "falied to start rpc response consumer")
// 		return nil, err
// 	}
// 	defer s.channel.Cancel(correlationId, false)

// 	publishing.CorrelationId = correlationId
// 	publishing.ReplyTo = s.cfg.RpcResponseQueue
// 	publishing.Type = operation

// 	if err = s.channel.PublishWithContext(ctx, "", queueName, false, false, publishing); err != nil {
// 		tracer.AddSpanErrorAndFail(span, err, "failed to publish rpc request")
// 		return nil, err
// 	}

// 	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.RpcTimeout)*time.Second)
// 	defer cancel()

// 	for {
// 		select {

// 		case <-ctx.Done():
// 			err := errors.New("rpc request timed out")
// 			tracer.FailSpan(span, err.Error())
// 			return nil, err

// 		case response := <-responses:
// 			if correlationId == response.CorrelationId {
// 				response.Ack(false)
// 				status, statusIsString := response.Headers["status"].(string)

// 				if !statusIsString {
// 					err := fmt.Errorf("RPC: %s returned non string response status", operation)
// 					tracer.FailSpan(span, err.Error())
// 					return nil, err
// 				}

// 				if status != "success" {
// 					err := errors.New(status)
// 					span.SetStatus(codes.Error, err.Error())
// 					return nil, err
// 				}

// 				return &response, nil
// 			}
// 		}
// 	}
// }

// // Requests a email to be queued for sending using the mailer service
// func (s *Server) RpcMailerService(ctx context.Context, publishing amqp.Publishing) (*amqp.Delivery, error) {
// 	return s.Rpc(ctx, s.cfg.MailerServiceQueue, "", publishing)
// }
