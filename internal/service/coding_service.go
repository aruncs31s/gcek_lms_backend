package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/aruncs/esdc-lms/pkg/coderunner"
	"github.com/google/uuid"
)

// CodingService handles all business logic for coding assignments and their submissions.
type CodingService interface {
	// Teacher / course creator operations
	CreateCodingAssignment(courseID, teacherID uuid.UUID, req *dto.CreateCodingAssignmentRequest) (*dto.CodingAssignmentResponse, error)
	GetCodingAssignmentsByCourse(courseID, requesterID uuid.UUID, isTeacher bool) ([]dto.CodingAssignmentResponse, error)
	GetCodingAssignmentByID(assignmentID, requesterID uuid.UUID, isTeacher bool) (*dto.CodingAssignmentResponse, error)
	UpdateCodingAssignment(assignmentID, teacherID uuid.UUID, req *dto.UpdateCodingAssignmentRequest) (*dto.CodingAssignmentResponse, error)
	DeleteCodingAssignment(assignmentID, teacherID uuid.UUID) error

	// Student operations
	RunCode(req *dto.RunCodeRequest) (*dto.RunCodeResponse, error)
	SubmitCode(assignmentID, studentID uuid.UUID, req *dto.SubmitCodingRequest) (*dto.CodingSubmissionResponse, error)
	GetMySubmission(assignmentID, studentID uuid.UUID) (*dto.CodingSubmissionResponse, error)

	// Teacher grading
	GetSubmissions(assignmentID, teacherID uuid.UUID) ([]dto.CodingSubmissionResponse, error)
	GradeSubmission(submissionID, teacherID uuid.UUID, req *dto.GradeCodingSubmissionRequest) (*dto.CodingSubmissionResponse, error)
}

type codingService struct {
	codingRepo  repository.CodingRepository
	courseRepo  repository.CourseRepository
}

func NewCodingService(codingRepo repository.CodingRepository, courseRepo repository.CourseRepository) CodingService {
	return &codingService{
		codingRepo: codingRepo,
		courseRepo: courseRepo,
	}
}

// ── Helper: auth checks ─────────────────────────────────────────────────────

func (s *codingService) checkTeacherAccess(courseID, userID uuid.UUID) error {
	course, err := s.courseRepo.GetCourseByID(courseID)
	if err != nil {
		return err
	}
	if course.TeacherID != userID {
		return errors.New("forbidden: only the course teacher can perform this action")
	}
	return nil
}

// ── Helper: marshal / unmarshal test cases ──────────────────────────────────

func marshalTestCases(cases []dto.TestCase) (string, error) {
	b, err := json.Marshal(cases)
	return string(b), err
}

func unmarshalTestCases(raw string) ([]dto.TestCase, error) {
	var cases []dto.TestCase
	if raw == "" || raw == "[]" {
		return cases, nil
	}
	err := json.Unmarshal([]byte(raw), &cases)
	return cases, err
}

func marshalTestResults(results []dto.TestResult) (string, error) {
	b, err := json.Marshal(results)
	return string(b), err
}

func unmarshalTestResults(raw string) ([]dto.TestResult, error) {
	var results []dto.TestResult
	if raw == "" || raw == "[]" {
		return results, nil
	}
	err := json.Unmarshal([]byte(raw), &results)
	return results, err
}

// ── Helper: map internal runner results → DTO ───────────────────────────────

func runnerResultsToDTO(runnerResults []coderunner.TestResult, cases []dto.TestCase, isTeacher bool) []dto.TestResult {
	// Build a map from ID → TestCase to quickly look up is_hidden flag
	caseMap := make(map[string]dto.TestCase, len(cases))
	for _, tc := range cases {
		caseMap[tc.ID] = tc
	}

	out := make([]dto.TestResult, 0, len(runnerResults))
	for _, rr := range runnerResults {
		tc := caseMap[rr.TestCaseID]
		tr := dto.TestResult{
			TestCaseID:      rr.TestCaseID,
			Description:     rr.Description,
			Actual:          rr.Actual,
			Passed:          rr.Passed,
			Error:           rr.Error,
			ExecutionTimeMs: rr.ExecutionTimeMs,
		}
		// Reveal input/expected only for non-hidden cases or teachers
		if !tc.IsHidden || isTeacher {
			tr.Input = rr.Input
			tr.Expected = rr.Expected
		}
		out = append(out, tr)
	}
	return out
}

// ── CodingAssignment CRUD ───────────────────────────────────────────────────

func (s *codingService) CreateCodingAssignment(courseID, teacherID uuid.UUID, req *dto.CreateCodingAssignmentRequest) (*dto.CodingAssignmentResponse, error) {
	if err := s.checkTeacherAccess(courseID, teacherID); err != nil {
		return nil, err
	}

	tcJSON, err := marshalTestCases(req.TestCases)
	if err != nil {
		return nil, err
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		parsed, err := time.Parse(time.RFC3339, req.DueDate)
		if err == nil {
			dueDate = &parsed
		}
	}

	a := &model.CodingAssignment{
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
		Language:    req.Language,
		StarterCode: req.StarterCode,
		TestCases:   tcJSON,
		MaxScore:    req.MaxScore,
		DueDate:     dueDate,
	}

	if err := s.codingRepo.CreateCodingAssignment(a); err != nil {
		return nil, err
	}

	return toCodingAssignmentResponse(a, req.TestCases, true), nil
}

func (s *codingService) GetCodingAssignmentsByCourse(courseID, _ uuid.UUID, isTeacher bool) ([]dto.CodingAssignmentResponse, error) {
	assignments, err := s.codingRepo.GetCodingAssignmentsByCoure(courseID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.CodingAssignmentResponse, 0, len(assignments))
	for i := range assignments {
		cases, _ := unmarshalTestCases(assignments[i].TestCases)
		responses = append(responses, *toCodingAssignmentResponse(&assignments[i], cases, isTeacher))
	}
	return responses, nil
}

func (s *codingService) GetCodingAssignmentByID(assignmentID, _ uuid.UUID, isTeacher bool) (*dto.CodingAssignmentResponse, error) {
	a, err := s.codingRepo.GetCodingAssignmentByID(assignmentID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("coding assignment not found")
	}

	cases, _ := unmarshalTestCases(a.TestCases)
	return toCodingAssignmentResponse(a, cases, isTeacher), nil
}

func (s *codingService) UpdateCodingAssignment(assignmentID, teacherID uuid.UUID, req *dto.UpdateCodingAssignmentRequest) (*dto.CodingAssignmentResponse, error) {
	a, err := s.codingRepo.GetCodingAssignmentByID(assignmentID)
	if err != nil || a == nil {
		return nil, errors.New("coding assignment not found")
	}

	if err := s.checkTeacherAccess(a.CourseID, teacherID); err != nil {
		return nil, err
	}

	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Description != nil {
		a.Description = *req.Description
	}
	if req.Language != nil {
		a.Language = *req.Language
	}
	if req.StarterCode != nil {
		a.StarterCode = *req.StarterCode
	}
	if req.TestCases != nil {
		tcJSON, err := marshalTestCases(req.TestCases)
		if err != nil {
			return nil, err
		}
		a.TestCases = tcJSON
	}
	if req.MaxScore != nil {
		a.MaxScore = *req.MaxScore
	}
	if req.DueDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err == nil {
			a.DueDate = &parsed
		}
	}

	if err := s.codingRepo.UpdateCodingAssignment(a); err != nil {
		return nil, err
	}

	cases, _ := unmarshalTestCases(a.TestCases)
	return toCodingAssignmentResponse(a, cases, true), nil
}

func (s *codingService) DeleteCodingAssignment(assignmentID, teacherID uuid.UUID) error {
	a, err := s.codingRepo.GetCodingAssignmentByID(assignmentID)
	if err != nil || a == nil {
		return errors.New("coding assignment not found")
	}
	if err := s.checkTeacherAccess(a.CourseID, teacherID); err != nil {
		return err
	}
	return s.codingRepo.DeleteCodingAssignment(assignmentID)
}

// ── Code execution ──────────────────────────────────────────────────────────

func (s *codingService) RunCode(req *dto.RunCodeRequest) (*dto.RunCodeResponse, error) {
	res, err := coderunner.RunCode(req.Language, req.Code, req.Input)
	if err != nil {
		return nil, err
	}

	return &dto.RunCodeResponse{
		Output:          res.Stdout,
		Stderr:          res.Stderr,
		Error:           res.Error,
		ExecutionTimeMs: res.ExecutionTimeMs,
	}, nil
}

// ── Submission ──────────────────────────────────────────────────────────────

func (s *codingService) SubmitCode(assignmentID, studentID uuid.UUID, req *dto.SubmitCodingRequest) (*dto.CodingSubmissionResponse, error) {
	a, err := s.codingRepo.GetCodingAssignmentByID(assignmentID)
	if err != nil || a == nil {
		return nil, errors.New("coding assignment not found")
	}

	cases, err := unmarshalTestCases(a.TestCases)
	if err != nil {
		return nil, err
	}

	// Convert DTO cases → runner cases
	runnerCases := make([]coderunner.TestCase, len(cases))
	for i, tc := range cases {
		runnerCases[i] = coderunner.TestCase{
			ID:             tc.ID,
			Description:    tc.Description,
			Input:          tc.Input,
			ExpectedOutput: tc.ExpectedOutput,
			IsHidden:       tc.IsHidden,
		}
	}

	// Run all test cases
	runnerResults := coderunner.RunTests(a.Language, req.Code, runnerCases)

	// Calculate score
	passed := 0
	for _, r := range runnerResults {
		if r.Passed {
			passed++
		}
	}
	var score *int
	if len(cases) > 0 {
		s_ := int(float64(passed) / float64(len(cases)) * float64(a.MaxScore))
		score = &s_
	}

	dtoResults := runnerResultsToDTO(runnerResults, cases, false)
	resultsJSON, _ := marshalTestResults(dtoResults)

	allPassed := passed == len(cases)

	// Check if there is an existing submission – update instead of duplicate
	existing, _ := s.codingRepo.GetCodingSubmissionByUser(assignmentID, studentID)
	if existing != nil {
		existing.Code = req.Code
		existing.Score = score
		existing.TestResults = resultsJSON
		existing.Passed = allPassed
		now := time.Now()
		existing.SubmittedAt = now
		existing.GradedAt = nil
		existing.Feedback = ""
		if err := s.codingRepo.UpdateCodingSubmission(existing); err != nil {
			return nil, err
		}
		return toCodingSubmissionResponse(existing, dtoResults, ""), nil
	}

	sub := &model.CodingSubmission{
		CodingAssignmentID: assignmentID,
		UserID:             studentID,
		Code:               req.Code,
		Score:              score,
		TestResults:        resultsJSON,
		Passed:             allPassed,
	}

	if err := s.codingRepo.CreateCodingSubmission(sub); err != nil {
		return nil, err
	}

	return toCodingSubmissionResponse(sub, dtoResults, ""), nil
}

func (s *codingService) GetMySubmission(assignmentID, studentID uuid.UUID) (*dto.CodingSubmissionResponse, error) {
	sub, err := s.codingRepo.GetCodingSubmissionByUser(assignmentID, studentID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, nil
	}
	results, _ := unmarshalTestResults(sub.TestResults)
	return toCodingSubmissionResponse(sub, results, ""), nil
}

func (s *codingService) GetSubmissions(assignmentID, teacherID uuid.UUID) ([]dto.CodingSubmissionResponse, error) {
	a, err := s.codingRepo.GetCodingAssignmentByID(assignmentID)
	if err != nil || a == nil {
		return nil, errors.New("coding assignment not found")
	}
	if err := s.checkTeacherAccess(a.CourseID, teacherID); err != nil {
		return nil, err
	}

	subs, err := s.codingRepo.GetCodingSubmissionsByAssignment(assignmentID)
	if err != nil {
		return nil, err
	}

	out := make([]dto.CodingSubmissionResponse, 0, len(subs))
	for i := range subs {
		results, _ := unmarshalTestResults(subs[i].TestResults)
		userName := ""
		if subs[i].User.Email != "" {
			userName = subs[i].User.Email
		}
		out = append(out, *toCodingSubmissionResponse(&subs[i], results, userName))
	}
	return out, nil
}

func (s *codingService) GradeSubmission(submissionID, teacherID uuid.UUID, req *dto.GradeCodingSubmissionRequest) (*dto.CodingSubmissionResponse, error) {
	sub, err := s.codingRepo.GetCodingSubmissionByID(submissionID)
	if err != nil || sub == nil {
		return nil, errors.New("submission not found")
	}

	a, err := s.codingRepo.GetCodingAssignmentByID(sub.CodingAssignmentID)
	if err != nil || a == nil {
		return nil, errors.New("coding assignment not found")
	}
	if err := s.checkTeacherAccess(a.CourseID, teacherID); err != nil {
		return nil, err
	}

	sub.Score = &req.Score
	sub.Feedback = req.Feedback
	now := time.Now()
	sub.GradedAt = &now

	if err := s.codingRepo.UpdateCodingSubmission(sub); err != nil {
		return nil, err
	}

	results, _ := unmarshalTestResults(sub.TestResults)
	return toCodingSubmissionResponse(sub, results, ""), nil
}

// ── Response mappers ────────────────────────────────────────────────────────

func toCodingAssignmentResponse(a *model.CodingAssignment, cases []dto.TestCase, isTeacher bool) *dto.CodingAssignmentResponse {
	// For students, strip hidden test case input/expected output
	visibleCases := make([]dto.TestCase, 0, len(cases))
	for _, tc := range cases {
		if tc.IsHidden && !isTeacher {
			visibleCases = append(visibleCases, dto.TestCase{
				ID:          tc.ID,
				Description: tc.Description,
				IsHidden:    true,
			})
		} else {
			visibleCases = append(visibleCases, tc)
		}
	}

	var dueDateStr *string
	if a.DueDate != nil {
		s := a.DueDate.Format(time.RFC3339)
		dueDateStr = &s
	}

	return &dto.CodingAssignmentResponse{
		ID:          a.ID,
		CourseID:    a.CourseID,
		Title:       a.Title,
		Description: a.Description,
		Language:    a.Language,
		StarterCode: a.StarterCode,
		TestCases:   visibleCases,
		MaxScore:    a.MaxScore,
		DueDate:     dueDateStr,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
	}
}

func toCodingSubmissionResponse(sub *model.CodingSubmission, results []dto.TestResult, userName string) *dto.CodingSubmissionResponse {
	var gradedAtStr *string
	if sub.GradedAt != nil {
		s := sub.GradedAt.Format(time.RFC3339)
		gradedAtStr = &s
	}

	return &dto.CodingSubmissionResponse{
		ID:                 sub.ID,
		CodingAssignmentID: sub.CodingAssignmentID,
		UserID:             sub.UserID,
		UserName:           userName,
		Code:               sub.Code,
		Score:              sub.Score,
		Feedback:           sub.Feedback,
		TestResults:        results,
		Passed:             sub.Passed,
		SubmittedAt:        sub.SubmittedAt.Format(time.RFC3339),
		GradedAt:           gradedAtStr,
	}
}
