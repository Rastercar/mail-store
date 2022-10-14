package queue

import (
	"log"
	"mail-store-ms/queue/interfaces"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpConnectionWrapper struct {
	conn *amqp.Connection
}

func (w AmqpConnectionWrapper) Close() error {
	return w.conn.Close()
}

func (w AmqpConnectionWrapper) Channel() (interfaces.AmqpChannel, error) {
	return w.conn.Channel()
}

type Connector struct{}

func (c *Connector) Connect(url string) (interfaces.AmqpConnection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return AmqpConnectionWrapper{conn}, nil
}

func (s *Server) connect() {
	currentAttempt := 1
	sleepTime := time.Second * time.Duration(s.cfg.ReconnectWaitTime)

	for {
		log.Printf("[ RMQ ] trying to connect, attempt: %d", currentAttempt)

		con, err := s.Connector.Connect(s.cfg.Url)

		if err != nil || con == nil {
			log.Printf("[ RMQ ] connection failed %v", err)

			currentAttempt++
			time.Sleep(sleepTime)

			continue
		}

		channel, err := con.Channel()
		if err != nil {
			log.Printf("[ RMQ ] connection channel failed %v", err)

			currentAttempt++
			time.Sleep(sleepTime)

			continue
		}

		_, err = channel.QueueDeclare(
			s.cfg.Queue, // name
			true,        // durable
			false,       // autodelete
			false,       // exclusive
			false,       // nowait
			nil,         // args
		)
		if err != nil {
			log.Fatalf("[ RMQ ] failed to declare exchange: %v ", err)
		}

		s.deliveries, err = channel.Consume(
			s.cfg.Queue, // queue
			"",          // consumer
			false,       // autoack
			false,       // exclusive
			false,       // nolocal
			false,       // nowait
			nil,         // args
		)
		if err != nil {
			log.Fatalf("[ RMQ ] failed to consume mail requests queue: %v ", err)
		}

		s.conn = con
		s.channel = channel

		s.notifyClose = make(chan *amqp.Error, 1024)
		s.channel.NotifyClose(s.notifyClose)

		s.startConsumer()

		log.Printf("[ RMQ ] connected")
		return
	}
}
