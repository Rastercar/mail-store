package db

import (
	"log"
	"mail-store-ms/db/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(dbUrl string, debugMode bool) *gorm.DB {
	config := gorm.Config{}

	if debugMode {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dbUrl), &config)
	if err != nil {
		log.Fatalf("[ DB ] failed to init db: %v", err)
	}

	if err = db.AutoMigrate(models.MailRequest{}, models.MailRequestFeedback{}); err != nil {
		log.Fatalf("[ DB ] failed to migrate db schema: %v", err)
	}

	return db
}
