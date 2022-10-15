package controller

import (
	"context"
	"encoding/json"
	"errors"
	"mail-store-ms/controller/dtos"
	"mail-store-ms/db/repository"
	"mail-store-ms/queue"
	"mail-store-ms/services/mail"
	"mail-store-ms/tracer"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Controller struct {
	repo      repository.MailRequestRepository
	server    *queue.Server
	mailer    mail.Mailer
	validator *validator.Validate
}

func NewRouter(server *queue.Server, mailer mail.Mailer, repo repository.MailRequestRepository, validator *validator.Validate) queue.RpcRouter {
	c := Controller{repo, server, mailer, validator}

	router := make(queue.RpcRouter)
	router["sendMail"] = c.sendMail
	router["__internal:mail-feeback__"] = c.registerMailRequestFeedback

	return router
}

func (c *Controller) registerMailRequestFeedback(ctx context.Context, d *amqp.Delivery) (res queue.RpcRes) {
	_, span := tracer.NewSpan(ctx, "controller", "registerMailRequestFeedback")
	defer span.End()

	if d.CorrelationId == "" {
		tracer.FailSpan(span, "cannot update mail request feedback with no correlation ID")
		return res
	}

	var dto mail.MailerServiceSendEmailRes

	if err := json.Unmarshal(d.Body, &dto); err != nil {
		errMsg := "failed to unmarshal mail request feedback"
		tracer.AddSpanErrorAndFail(span, err, errMsg)
		return res
	}

	if dto.Success {
		err := c.repo.UpdateMailRequestFeedback(d.CorrelationId, &dto.Timestamp)
		span.RecordError(err)
	} else {
		err := c.repo.UpdateMailRequestFeedback(d.CorrelationId, nil)
		span.RecordError(err)
	}

	return res
}

func (c *Controller) sendMail(ctx context.Context, d *amqp.Delivery) queue.RpcRes {
	ctx, span := tracer.NewSpan(ctx, "controller", "sendMail")
	defer span.End()

	var dto dtos.SendEmailDto

	if err := json.Unmarshal(d.Body, &dto); err != nil {
		errMsg := "failed to unmarshal send mail request"

		tracer.AddSpanErrorAndFail(span, err, errMsg)
		return queue.RpcRes{Error: errors.New(errMsg)}
	}

	if err := c.validator.Struct(dto); err != nil {
		tracer.AddSpanErrorAndFail(span, err, "invalid request body")
		return queue.RpcRes{Error: err}
	}

	mailUuid := uuid.NewString()

	err := c.mailer.SendEmail(ctx, dto, mailUuid)
	if err != nil {
		tracer.AddSpanErrorAndFail(span, err, "failed to send email")
		return queue.RpcRes{Error: err}
	}

	response, _ := json.Marshal(dtos.SendEmailRes{Uuid: mailUuid})

	return queue.RpcRes{ResponseBody: response}
}
