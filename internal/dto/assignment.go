package dto

import "github.com/google/uuid"

type CreateAssignmentRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	MaxScore    int    `json:"max_score" binding:"required"`
	DueDate     string `json:"due_date"` // Optional ISO8601 string
}

type UpdateAssignmentRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	MaxScore    *int    `json:"max_score"`
	DueDate     *string `json:"due_date"`
}

type AssignmentResponse struct {
	ID          uuid.UUID `json:"id"`
	CourseID    uuid.UUID `json:"course_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	MaxScore    int       `json:"max_score"`
	DueDate     *string   `json:"due_date,omitempty"`
	CreatedAt   string    `json:"created_at"`
}

type SubmitAssignmentRequest struct {
	FileURL string `json:"file_url" binding:"required"`
}

type GradeSubmissionRequest struct {
	Score    int    `json:"score" binding:"required"`
	Feedback string `json:"feedback"`
}

type AssignmentSubmissionResponse struct {
	ID            uuid.UUID `json:"id"`
	AssignmentID  uuid.UUID `json:"assignment_id"`
	UserID        uuid.UUID `json:"user_id"`
	UserName      string    `json:"user_name,omitempty"`
	FileURL       string    `json:"file_url"`
	ExtractedText string    `json:"extracted_text"`
	Score         *int      `json:"score,omitempty"`
	Feedback      string    `json:"feedback,omitempty"`
	SubmittedAt   string    `json:"submitted_at"`
	GradedAt      *string   `json:"graded_at,omitempty"`
}
