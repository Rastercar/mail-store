package repository

import (
	"mail-store-ms/db/models"
	"time"

	"gorm.io/gorm"
)

type MailerRepo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *MailerRepo {
	return &MailerRepo{db: db}
}

func (r *MailerRepo) StoreMailRequest(req models.MailRequest) (mailReqId int, err error) {
	err = r.db.Create(&req).Error
	return req.Id, err
}

func (r *MailerRepo) StoreMailRequestFeedback(req models.MailRequestFeedback) error {
	return r.db.Create(&req).Error
}

func (r *MailerRepo) UpdateMailRequestFeedback(uuid string, queuedAt *time.Time) error {
	success := queuedAt != nil
	return r.db.Exec("UPDATE mail_request_feedback SET success = ?, queued_at = ? WHERE uuid = ?", success, queuedAt, uuid).Error
}
