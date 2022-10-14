package repository

import (
	"mail-store-ms/db/models"

	"gorm.io/gorm"
)

type MailerRepo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *MailerRepo {
	return &MailerRepo{db: db}
}

func (r *MailerRepo) StoreMailRequest(req models.MailRequest) error {
	return r.db.Create(&req).Error
}
