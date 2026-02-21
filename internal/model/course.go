package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Course struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TeacherID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Title        string    `gorm:"type:varchar(255);not null"`
	Description  string    `gorm:"type:text"`
	ThumbnailURL string    `gorm:"type:varchar(255)"`
	Price        float64   `gorm:"type:decimal(10,2);default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	Modules     []Module     `gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Enrollments []Enrollment `gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (c *Course) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}

type Module struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CourseID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Title      string    `gorm:"type:varchar(255);not null"`
	VideoURL   string    `gorm:"type:varchar(500)"` // Optional or required based on business rule
	OrderIndex int       `gorm:"type:int;not null;default:0"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (m *Module) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

type EnrollmentStatus string

const (
	EnrollmentStatusActive    EnrollmentStatus = "active"
	EnrollmentStatusCompleted EnrollmentStatus = "completed"
)

type Enrollment struct {
	UserID             uuid.UUID        `gorm:"type:uuid;primaryKey;autoIncrement:false"`
	CourseID           uuid.UUID        `gorm:"type:uuid;primaryKey;autoIncrement:false"`
	Status             EnrollmentStatus `gorm:"type:varchar(50);default:'active'"`
	ProgressPercentage float64          `gorm:"type:float;default:0"`
	EnrolledAt         time.Time        `gorm:"autoCreateTime"`
}
