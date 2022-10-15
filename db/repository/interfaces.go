package repository

import (
	"mail-store-ms/db/models"
	"time"
)

type MailRequestRepository interface {
	StoreMailRequest(models.MailRequest) (mailReqId int, err error)
	UpdateMailRequestFeedback(uuid string, queuedAt *time.Time) error
	StoreMailRequestFeedback(models.MailRequestFeedback) error
}
