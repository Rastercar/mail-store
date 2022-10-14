package mail

import (
	"context"
	"encoding/json"
	"mail-store-ms/config"
	"mail-store-ms/controller/dtos"
	"mail-store-ms/queue"
	"mail-store-ms/tracer"
	"mail-store-ms/utils/arrays"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Mailer struct {
	cfg   *config.Config
	queue *queue.Server
}

func New(cfg *config.Config, queue *queue.Server) Mailer {
	return Mailer{cfg, queue}
}

const maxSesRecipients = 50

// Send email requests a email to be sent by the mailer service, sending as many
// requests are needed to send the email to all recipients.
func (m *Mailer) SendEmail(ctx context.Context, email dtos.SendEmailDto, uuid string) error {
	ctx, span := tracer.NewSpan(ctx, "mail", "SendEmail")
	defer span.End()

	recipientChunks := arrays.ToChunks(arrays.RemoveDuplicates(email.To), maxSesRecipients)

	for _, recipientChunk := range recipientChunks {
		email.To = recipientChunk

		body, err := json.Marshal(email)
		if err != nil {
			tracer.AddSpanErrorAndFail(span, err, "failed to marshal mail request for mailer service")

			// since the DTO is nearly the same except the 'To' property we assume the marshaling will
			// always fail so return the error
			return err
		}

		err = m.queue.Publish(ctx, "", m.cfg.Rmq.MailerServiceQueue, amqp.Publishing{
			Body:          body,
			CorrelationId: uuid,
		})
		if err != nil {
			span.RecordError(err)
		}
	}

	return nil
}
