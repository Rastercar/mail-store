package repository

import "mail-store-ms/db/models"

type MailRequestRepository interface {
	StoreMailRequest(models.MailRequest) error
}
