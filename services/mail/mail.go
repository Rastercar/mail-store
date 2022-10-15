package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"mail-store-ms/config"
	"mail-store-ms/controller/dtos"
	"mail-store-ms/db/models"
	"mail-store-ms/db/repository"
	"mail-store-ms/queue"
	"mail-store-ms/tracer"
	"mail-store-ms/utils/arrays"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Mailer struct {
	cfg   *config.Config
	queue *queue.Server
	repo  repository.MailRequestRepository
}

func New(cfg *config.Config, queue *queue.Server, repo repository.MailRequestRepository) Mailer {
	return Mailer{cfg, queue, repo}
}

const maxSesRecipients = 50

// Send email requests a email to be sent by the mailer service, sending as many
// requests are needed to send the email to all recipients.
func (m *Mailer) SendEmail(ctx context.Context, email dtos.SendEmailDto, mailUuid string) error {
	ctx, span := tracer.NewSpan(ctx, "mail", "SendEmail")
	defer span.End()

	recipientChunks := arrays.ToChunks(arrays.RemoveDuplicates(email.To), maxSesRecipients)

	mailerServiceDto := MailerServiceSendEmailDto{
		Uuid:             mailUuid,
		ReplyToAddresses: email.ReplyToAddresses,
		SubjectText:      email.SubjectText,
		BodyHtml:         email.BodyHtml,
		BodyText:         email.BodyText,
	}

	_, err := json.Marshal(mailerServiceDto)
	if err != nil {
		tracer.AddSpanErrorAndFail(span, err, "failed to marshal mail request for mailer service")
		// since the DTO is nearly the same except the 'To' property we assume the marshaling will
		// always fail so return the error
		return err
	}

	mailRequest := models.MailRequest{
		Uuid:             mailUuid,
		To:               email.To,
		ReplyToAddresses: email.ReplyToAddresses,
		SubjectText:      email.SubjectText,
		BodyText:         email.BodyText,
		BodyHtml:         email.BodyHtml,
	}

	mailReqId, err := m.repo.StoreMailRequest(mailRequest)
	if err != nil {
		tracer.AddSpanErrorAndFail(span, err, "internal db error storing mail_request")
		return err
	}

	for _, recipientChunk := range recipientChunks {
		mailerServiceDto.To = recipientChunk

		// Here we can safely ignore the error as the marshal was tested above
		// and the change in the 'To' property will never cause a marshaling failure
		body, _ := json.Marshal(mailerServiceDto)

		requestUuid := uuid.NewString()

		publishing := amqp.Publishing{
			Body:          body,
			CorrelationId: requestUuid,
			ReplyTo:       m.cfg.Rmq.MailerServiceResponseQueue,
		}

		if err = m.queue.Publish(ctx, "", m.cfg.Rmq.MailerServiceQueue, publishing); err != nil {
			span.RecordError(fmt.Errorf("failed to publish mail req: %v", err))
		}

		mailReqFeedback := models.MailRequestFeedback{
			Uuid:          requestUuid,
			Recipients:    recipientChunk,
			MailRequestId: mailReqId,
			Success:       nil,
		}

		if err = m.repo.StoreMailRequestFeedback(mailReqFeedback); err != nil {
			span.RecordError(fmt.Errorf("failed to store mail req feedback: %v", err))
		}
	}

	return nil
}
