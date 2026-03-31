package dto

import "time"

type GenerateCertificateRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	CourseID string `json:"course_id" validate:"required"`
	// Layout selects one of the 10 certificate designs (1-10). Defaults to 1.
	Layout int `json:"layout"`
}

type CertificateResponse struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	CourseID string    `json:"course_id"`
	FileURL  string    `json:"file_url"`
	IssuedAt time.Time `json:"issued_at"`
	Layout   int       `json:"layout"`
}

