package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CodingAssignment represents a programming-based assignment inside a course.
// Each assignment can have multiple test cases stored as a JSON array in TestCases.
type CodingAssignment struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CourseID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	Language    string    `gorm:"type:varchar(32);not null;default:'python'"` // python | javascript
	StarterCode string    `gorm:"type:text"`                                  // Code skeleton given to the student
	TestCases   string    `gorm:"type:jsonb;not null;default:'[]'"`            // JSON array of TestCase objects
	MaxScore    int       `gorm:"default:100"`
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Course      Course              `gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Submissions []CodingSubmission  `gorm:"foreignKey:CodingAssignmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (ca *CodingAssignment) BeforeCreate(tx *gorm.DB) (err error) {
	if ca.ID == uuid.Nil {
		ca.ID = uuid.New()
	}
	return
}

// CodingSubmission stores a student's code submission for a CodingAssignment.
type CodingSubmission struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CodingAssignmentID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID             uuid.UUID `gorm:"type:uuid;not null;index"`
	Code               string    `gorm:"type:text;not null"`
	Score              *int      `gorm:"type:int"`
	Feedback           string    `gorm:"type:text"`
	TestResults        string    `gorm:"type:jsonb;not null;default:'[]'"` // JSON array of TestResult objects
	Passed             bool      `gorm:"default:false"`
	SubmittedAt        time.Time `gorm:"autoCreateTime"`
	GradedAt           *time.Time

	CodingAssignment CodingAssignment `gorm:"foreignKey:CodingAssignmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User             User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (cs *CodingSubmission) BeforeCreate(tx *gorm.DB) (err error) {
	if cs.ID == uuid.Nil {
		cs.ID = uuid.New()
	}
	return
}
