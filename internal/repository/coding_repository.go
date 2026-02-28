package repository

import (
	"errors"

	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CodingRepository interface {
	// Assignments
	CreateCodingAssignment(a *model.CodingAssignment) error
	GetCodingAssignmentsByCoure(courseID uuid.UUID) ([]model.CodingAssignment, error)
	GetCodingAssignmentByID(id uuid.UUID) (*model.CodingAssignment, error)
	UpdateCodingAssignment(a *model.CodingAssignment) error
	DeleteCodingAssignment(id uuid.UUID) error

	// Submissions
	CreateCodingSubmission(s *model.CodingSubmission) error
	GetCodingSubmissionsByAssignment(assignmentID uuid.UUID) ([]model.CodingSubmission, error)
	GetCodingSubmissionByUser(assignmentID, userID uuid.UUID) (*model.CodingSubmission, error)
	GetCodingSubmissionByID(id uuid.UUID) (*model.CodingSubmission, error)
	UpdateCodingSubmission(s *model.CodingSubmission) error
}

type codingRepository struct {
	db *gorm.DB
}

func NewCodingRepository(db *gorm.DB) CodingRepository {
	return &codingRepository{db: db}
}

// ── Coding Assignments ──────────────────────────────────────────────────────

func (r *codingRepository) CreateCodingAssignment(a *model.CodingAssignment) error {
	return r.db.Create(a).Error
}

func (r *codingRepository) GetCodingAssignmentsByCoure(courseID uuid.UUID) ([]model.CodingAssignment, error) {
	var assignments []model.CodingAssignment
	err := r.db.Where("course_id = ?", courseID).Order("created_at ASC").Find(&assignments).Error
	return assignments, err
}

func (r *codingRepository) GetCodingAssignmentByID(id uuid.UUID) (*model.CodingAssignment, error) {
	var a model.CodingAssignment
	err := r.db.First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &a, err
}

func (r *codingRepository) UpdateCodingAssignment(a *model.CodingAssignment) error {
	return r.db.Save(a).Error
}

func (r *codingRepository) DeleteCodingAssignment(id uuid.UUID) error {
	return r.db.Delete(&model.CodingAssignment{}, "id = ?", id).Error
}

// ── Coding Submissions ──────────────────────────────────────────────────────

func (r *codingRepository) CreateCodingSubmission(s *model.CodingSubmission) error {
	return r.db.Create(s).Error
}

func (r *codingRepository) GetCodingSubmissionsByAssignment(assignmentID uuid.UUID) ([]model.CodingSubmission, error) {
	var subs []model.CodingSubmission
	err := r.db.Preload("User").Where("coding_assignment_id = ?", assignmentID).Find(&subs).Error
	return subs, err
}

func (r *codingRepository) GetCodingSubmissionByUser(assignmentID, userID uuid.UUID) (*model.CodingSubmission, error) {
	var s model.CodingSubmission
	err := r.db.Where("coding_assignment_id = ? AND user_id = ?", assignmentID, userID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *codingRepository) GetCodingSubmissionByID(id uuid.UUID) (*model.CodingSubmission, error) {
	var s model.CodingSubmission
	err := r.db.Preload("User").First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *codingRepository) UpdateCodingSubmission(s *model.CodingSubmission) error {
	return r.db.Save(s).Error
}
