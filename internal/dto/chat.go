package dto

import "time"

type CreateConversationRequest struct {
	Type           string   `json:"type" validate:"required,oneof=direct group"`
	ParticipantIDs []string `json:"participant_ids" validate:"required,min=1"`
}

type ConversationResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type SendMessageRequest struct {
	Content string `json:"content" validate:"required"`
}

type MessageResponse struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Content        string    `json:"content"`
	AttachmentURL  *string   `json:"attachment_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
