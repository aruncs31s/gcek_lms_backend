package model

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents an in-app system or activity notification for a user
type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	Title     string    `gorm:"size:255;not null"`
	Message   string    `gorm:"type:text;not null"`
	IsRead    bool      `gorm:"default:false"`
	Type      string    `gorm:"size:50;not null"` // e.g. "assignment_submitted", "assignment_graded", "system"
	CreatedAt time.Time `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID"`
}
