package queue

import (
	"context"
	"fmt"
	"mail-store-ms/tracer"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (s *Server) consumeDeliveries() {
	for d := range s.deliveries {
		go s.ServeDelivery(&d)
	}
}

func (s *Server) consumeMailResponses() {
	for d := range s.mailRequestResponses {
		d.Type = "__internal:mail-feeback__"
		d.ReplyTo = "" // assure the server wont send a response

		go s.ServeDelivery(&d)
	}
}

// Routes a amqp delivery to its correct handler based on the delivery "Type"
// property, assuming its the name of the method to invoke, responses are delivered
// to the default exchanged with the message "ReplyTo" property as the routing key,
// so theyre delivered to the queue named in it.
//
// The response status is sent on the message "Type" property, any value different
// than "success" should be treated as an error description, in such case the message
// body *might* contain additional data about the error, otherwise the body is the
// success case response.
func (s *Server) ServeDelivery(d *amqp.Delivery) {
	ctx := tracer.ExtractAMPQHeaders(context.Background(), d.Headers)

	ctx, span := tracer.NewSpan(ctx, "amqp", fmt.Sprintf("AMQP - %s", d.Type))
	defer span.End()

	callHandler, routerHasCallHandler := s.DeliveryRouter[d.Type]
	hasReplyQueue := d.ReplyTo != ""

	if !routerHasCallHandler {
		d.Reject(false)

		if hasReplyQueue {
			s.Publish(ctx, "", d.ReplyTo, amqp.Publishing{
				Headers:       amqp.Table{"status": "unregistered handler"},
				CorrelationId: d.CorrelationId,
			})
		}

		tracer.AddSpanErrorAndFail(span, fmt.Errorf("unregistered handler: %s", d.Type), "unknown request type")
		return
	}

	res := callHandler(ctx, d)

	if res.Error == nil {
		tracer.RecordSpanSuccess(span, "message proccessed")
		d.Ack(false)
	} else {
		tracer.AddSpanErrorAndFail(span, res.Error, "error processing request")
		d.Reject(false)
	}

	if hasReplyQueue {
		s.Publish(ctx, "", d.ReplyTo, amqp.Publishing{
			CorrelationId: d.CorrelationId,
			Body:          res.ResponseBody,
			Headers:       amqp.Table{"status": "success"},
		})
	}
}
