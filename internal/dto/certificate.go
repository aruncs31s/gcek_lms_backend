package dto

import "time"

type GenerateCertificateRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	CourseID string `json:"course_id" validate:"required"`
}

type CertificateResponse struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	CourseID string    `json:"course_id"`
	FileURL  string    `json:"file_url"`
	IssuedAt time.Time `json:"issued_at"`
}
