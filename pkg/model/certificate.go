package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Certificate struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;index"`
	CourseID uuid.UUID `gorm:"type:uuid;not null;index"`
	FileURL  string    `gorm:"type:varchar(255);not null"`
	IssuedAt time.Time `gorm:"autoCreateTime"`

	User   User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Course Course `gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// To prevent multiple certificates per user-course
	// We might consider a unique index, but depends on product rules
	// _       struct{}  `gorm:"uniqueIndex:idx_user_course"`
}

func (c *Certificate) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
