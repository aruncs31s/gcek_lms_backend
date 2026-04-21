package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CourseReview struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CourseID  uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Rating    int       `gorm:"type:int;not null;check:rating >= 1 AND rating <= 5"`
	Comment   string    `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (cr *CourseReview) BeforeCreate(tx *gorm.DB) (err error) {
	if cr.ID == uuid.Nil {
		cr.ID = uuid.New()
	}
	return
}

type CourseLike struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;autoIncrement:false"`
	CourseID  uuid.UUID `gorm:"type:uuid;primaryKey;autoIncrement:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type WatchLater struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;autoIncrement:false"`
	CourseID  uuid.UUID `gorm:"type:uuid;primaryKey;autoIncrement:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
