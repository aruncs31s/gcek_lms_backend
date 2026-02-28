package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct"
	ConversationTypeGroup  ConversationType = "group"
)

type Conversation struct {
	ID        uuid.UUID        `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Type      ConversationType `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Participants []ConversationParticipant `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Messages     []Message                 `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}

type ConversationParticipant struct {
	ConversationID uuid.UUID `gorm:"type:uuid;primaryKey;autoIncrement:false"`
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey;autoIncrement:false"`
	JoinedAt       time.Time `gorm:"autoCreateTime"`
}

type Message struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;index"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Content        string    `gorm:"type:text;not null"`
	AttachmentURL  *string   `gorm:"type:varchar(255)"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
