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
	Type         string    `gorm:"type:varchar(50);default:'paid'"`        // free, paid
	Status       string    `gorm:"type:varchar(50);default:'not started'"` // not started, ended, active
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	Teacher     User         `gorm:"foreignKey:TeacherID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
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
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CourseID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index"` // Nullable for top-level modules
	Title       string     `gorm:"size:255;not null"`
	Description string     `gorm:"type:text"`
	Type        string     `gorm:"size:50;not null;default:'video'"` // "video" or "chapter"
	VideoURL    string
	Points      int  `gorm:"default:0"`
	IsFree      bool `gorm:"default:false"`
	OrderIndex  int  `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (m *Module) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

type ModuleProgress struct {
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey;not null"`
	ModuleID    uuid.UUID `gorm:"type:uuid;primaryKey;not null"`
	Completed   bool      `gorm:"default:false"`
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

	Course Course `gorm:"foreignKey:CourseID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
