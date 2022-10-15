package models

import (
	"database/sql"

	"github.com/lib/pq"
)

// A mail sending request, as sent by the client.
// Since mail requests can contain 1 or thousands of recipients
// a mail request may need many send operations to be completed.
type MailRequest struct {
	Id               int            `json:"id" gorm:"primaryKey"`                  //
	Uuid             string         `json:"uuid" gorm:"type:uuid;not null;unique"` // The externail ID used by clients/other systems to identify the email
	To               pq.StringArray `json:"to" gorm:"type:text[]"`                 //
	Cc               pq.StringArray `json:"cc" gorm:"type:text[]"`                 //
	Bcc              pq.StringArray `json:"bcc" gorm:"type:text[]"`                //
	ReplyToAddresses pq.StringArray `json:"reply_to_adresses" gorm:"type:text[]"`  //
	SubjectText      string         `json:"subject_text"`                          //
	BodyText         string         `json:"body_text"`                             //
	BodyHtml         string         `json:"body_html"`                             //

	//
	Feedbacks []MailRequestFeedback `gorm:"foreignKey:MailRequestId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (MailRequest) TableName() string {
	return "mail_request"
}

// feedback regarding one of possibly many mail sending operations
// needed to send the email to all recipients of a mail sending request
type MailRequestFeedback struct {
	Id            int            `json:"id" gorm:"primaryKey"`                  // the email sent by the operation
	Uuid          string         `json:"uuid" gorm:"type:uuid;not null;unique"` // uuid of the operation, NOT of the mail request
	Recipients    pq.StringArray `json:"recipients" gorm:"type:text[]"`         // all the recipients the email was sent to in this op
	Success       *bool          `json:"success"`                               // if the operation was a success, null means no response has been recieved
	QueuedAt      sql.NullTime   `json:"queued_at"`                             // when the email was successfully queued, null if operation failed or hasnt ended
	MailRequestId int            `json:"mail_request_id"`
	MailRequest   MailRequest    `gorm:"foreignKey:MailRequestId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (MailRequestFeedback) TableName() string {
	return "mail_request_feedback"
}
