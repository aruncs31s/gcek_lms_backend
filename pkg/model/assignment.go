package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Assignment struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CourseID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	MaxScore    int       `gorm:"default:100"`
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Course      Course                 `gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Submissions []AssignmentSubmission `gorm:"foreignKey:AssignmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (a *Assignment) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}

type AssignmentSubmission struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AssignmentID  uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	FileURL       string    `gorm:"type:varchar(255);not null"`
	ExtractedText string    `gorm:"type:text"` // Output from OCR service
	Score         *int      `gorm:"type:int"`  // Nullable, set by teacher
	Feedback      string    `gorm:"type:text"` // Teacher's feedback
	SubmittedAt   time.Time `gorm:"autoCreateTime"`
	GradedAt      *time.Time

	Assignment Assignment `gorm:"foreignKey:AssignmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User       User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (as *AssignmentSubmission) BeforeCreate(tx *gorm.DB) (err error) {
	if as.ID == uuid.Nil {
		as.ID = uuid.New()
	}
	return
}
