package repository

import (
	"errors"

	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssignmentRepository interface {
	CreateAssignment(assignment *model.Assignment) error
	GetAssignmentsByCourseID(courseID uuid.UUID) ([]model.Assignment, error)
	GetAssignmentByID(assignmentID uuid.UUID) (*model.Assignment, error)
	UpdateAssignment(assignment *model.Assignment) error
	DeleteAssignment(assignmentID uuid.UUID) error
	CreateSubmission(submission *model.AssignmentSubmission) error
	GetSubmissionsByAssignmentID(assignmentID uuid.UUID) ([]model.AssignmentSubmission, error)
	GetSubmissionByID(submissionID uuid.UUID) (*model.AssignmentSubmission, error)
	GetSubmissionByUserAndAssignment(userID, assignmentID uuid.UUID) (*model.AssignmentSubmission, error)
	UpdateSubmission(submission *model.AssignmentSubmission) error
}

type assignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &assignmentRepository{db: db}
}

func (r *assignmentRepository) CreateAssignment(assignment *model.Assignment) error {
	return r.db.Create(assignment).Error
}

func (r *assignmentRepository) GetAssignmentsByCourseID(courseID uuid.UUID) ([]model.Assignment, error) {
	var assignments []model.Assignment
	err := r.db.Where("course_id = ?", courseID).Order("created_at desc").Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepository) GetAssignmentByID(assignmentID uuid.UUID) (*model.Assignment, error) {
	var assignment model.Assignment
	err := r.db.Where("id = ?", assignmentID).First(&assignment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("assignment not found")
		}
		return nil, err
	}
	return &assignment, nil
}

func (r *assignmentRepository) UpdateAssignment(assignment *model.Assignment) error {
	return r.db.Save(assignment).Error
}

func (r *assignmentRepository) DeleteAssignment(assignmentID uuid.UUID) error {
	return r.db.Delete(&model.Assignment{}, "id = ?", assignmentID).Error
}

func (r *assignmentRepository) CreateSubmission(submission *model.AssignmentSubmission) error {
	return r.db.Create(submission).Error
}

func (r *assignmentRepository) GetSubmissionsByAssignmentID(assignmentID uuid.UUID) ([]model.AssignmentSubmission, error) {
	var submissions []model.AssignmentSubmission
	err := r.db.Preload("User").Preload("User.Profile").Where("assignment_id = ?", assignmentID).Order("submitted_at desc").Find(&submissions).Error
	return submissions, err
}

func (r *assignmentRepository) GetSubmissionByID(submissionID uuid.UUID) (*model.AssignmentSubmission, error) {
	var submission model.AssignmentSubmission
	err := r.db.Where("id = ?", submissionID).First(&submission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("submission not found")
		}
		return nil, err
	}
	return &submission, nil
}

func (r *assignmentRepository) GetSubmissionByUserAndAssignment(userID, assignmentID uuid.UUID) (*model.AssignmentSubmission, error) {
	var submission model.AssignmentSubmission
	err := r.db.Where("user_id = ? AND assignment_id = ?", userID, assignmentID).First(&submission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil to indicate no submission found without erroring out
		}
		return nil, err
	}
	return &submission, nil
}

func (r *assignmentRepository) UpdateSubmission(submission *model.AssignmentSubmission) error {
	return r.db.Save(submission).Error
}
