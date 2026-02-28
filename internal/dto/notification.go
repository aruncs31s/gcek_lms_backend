package dto

import "github.com/google/uuid"

type NotificationResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	Type      string    `json:"type"`
	CreatedAt string    `json:"created_at"`
}
