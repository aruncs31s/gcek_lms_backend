package dto

import "github.com/google/uuid"

// ──────────────────────────────────────────────
// Test-case definitions (shared between req/res)
// ──────────────────────────────────────────────

// TestCase defines a single test case for a coding assignment.
// Input is the raw stdin string passed to the program.
// ExpectedOutput is what stdout should equal (after trimming).
// IsHidden = true means the student cannot see this test case.
type TestCase struct {
	ID             string `json:"id"`
	Description    string `json:"description"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	IsHidden       bool   `json:"is_hidden"`
}

// TestResult is the outcome of running one test case against submitted code.
type TestResult struct {
	TestCaseID     string `json:"test_case_id"`
	Description    string `json:"description"`
	Input          string `json:"input,omitempty"`   // omitted for hidden cases
	Expected       string `json:"expected,omitempty"` // omitted for hidden cases
	Actual         string `json:"actual"`
	Passed         bool   `json:"passed"`
	Error          string `json:"error,omitempty"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
}

// ──────────────────────────────────────────────
// Coding Assignment CRUD
// ──────────────────────────────────────────────

type CreateCodingAssignmentRequest struct {
	Title       string     `json:"title"       binding:"required"`
	Description string     `json:"description"`
	Language    string     `json:"language"    binding:"required,oneof=python javascript"`
	StarterCode string     `json:"starter_code"`
	TestCases   []TestCase `json:"test_cases"  binding:"required"`
	MaxScore    int        `json:"max_score"   binding:"required"`
	DueDate     string     `json:"due_date"` // ISO 8601 optional
}

type UpdateCodingAssignmentRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Language    *string    `json:"language"`
	StarterCode *string    `json:"starter_code"`
	TestCases   []TestCase `json:"test_cases"`
	MaxScore    *int       `json:"max_score"`
	DueDate     *string    `json:"due_date"`
}

type CodingAssignmentResponse struct {
	ID          uuid.UUID  `json:"id"`
	CourseID    uuid.UUID  `json:"course_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Language    string     `json:"language"`
	StarterCode string     `json:"starter_code"`
	TestCases   []TestCase `json:"test_cases"` // hidden test cases stripped for students
	MaxScore    int        `json:"max_score"`
	DueDate     *string    `json:"due_date,omitempty"`
	CreatedAt   string     `json:"created_at"`
}

// ──────────────────────────────────────────────
// Run (ephemeral, not stored)
// ──────────────────────────────────────────────

// RunCodeRequest lets a student execute code without submitting.
type RunCodeRequest struct {
	Code     string `json:"code"     binding:"required"`
	Language string `json:"language" binding:"required,oneof=python javascript"`
	Input    string `json:"input"`   // arbitrary stdin for quick testing
}

type RunCodeResponse struct {
	Output          string `json:"output"`
	Stderr          string `json:"stderr,omitempty"`
	Error           string `json:"error,omitempty"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
}

// ──────────────────────────────────────────────
// Submission
// ──────────────────────────────────────────────

type SubmitCodingRequest struct {
	Code string `json:"code" binding:"required"`
}

type CodingSubmissionResponse struct {
	ID                 uuid.UUID    `json:"id"`
	CodingAssignmentID uuid.UUID    `json:"coding_assignment_id"`
	UserID             uuid.UUID    `json:"user_id"`
	UserName           string       `json:"user_name,omitempty"`
	Code               string       `json:"code"`
	Score              *int         `json:"score,omitempty"`
	Feedback           string       `json:"feedback,omitempty"`
	TestResults        []TestResult `json:"test_results"`
	Passed             bool         `json:"passed"`
	SubmittedAt        string       `json:"submitted_at"`
	GradedAt           *string      `json:"graded_at,omitempty"`
}

// ──────────────────────────────────────────────
// Grading
// ──────────────────────────────────────────────

type GradeCodingSubmissionRequest struct {
	Score    int    `json:"score"    binding:"required"`
	Feedback string `json:"feedback"`
}
