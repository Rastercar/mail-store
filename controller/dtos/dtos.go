package dtos

type SendEmailDto struct {
	To               []string `json:"to" validate:"dive,email"`
	ReplyToAddresses []string `json:"reply_to_addresses" validate:"dive,email"`
	SubjectText      string   `json:"subject_text" validate:"required"`
	BodyText         string   `json:"body_text"`
	BodyHtml         string   `json:"body_html" validate:"required"`
}

type SendEmailRes struct {
	Uuid string `json:"uuid"`
}
