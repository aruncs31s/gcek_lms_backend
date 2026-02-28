package service

import (
	"errors"
	"net/url"
	"time"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/aruncs/esdc-lms/pkg/ocr"
	"github.com/google/uuid"
)

type AssignmentService interface {
	CreateAssignment(courseID, teacherID uuid.UUID, req *dto.CreateAssignmentRequest) (*dto.AssignmentResponse, error)
	GetAssignmentsByCourse(courseID, userID uuid.UUID) ([]dto.AssignmentResponse, error)
	GetAssignmentByID(courseID, assignmentID, userID uuid.UUID) (*dto.AssignmentResponse, error)
	UpdateAssignment(courseID, assignmentID, teacherID uuid.UUID, req *dto.UpdateAssignmentRequest) (*dto.AssignmentResponse, error)
	DeleteAssignment(courseID, assignmentID, teacherID uuid.UUID) error
	SubmitAssignment(courseID, assignmentID, studentID uuid.UUID, req *dto.SubmitAssignmentRequest) (*dto.AssignmentSubmissionResponse, error)
	GetSubmissions(courseID, assignmentID, teacherID uuid.UUID) ([]dto.AssignmentSubmissionResponse, error)
	GradeSubmission(courseID, assignmentID, submissionID, teacherID uuid.UUID, req *dto.GradeSubmissionRequest) (*dto.AssignmentSubmissionResponse, error)
	GetStudentSubmission(courseID, assignmentID, studentID uuid.UUID) (*dto.AssignmentSubmissionResponse, error)
}

type assignmentService struct {
	assignmentRepo      repository.AssignmentRepository
	courseRepo          repository.CourseRepository
	ocrClient           *ocr.Client
	notificationService NotificationService
}

func NewAssignmentService(assignmentRepo repository.AssignmentRepository, courseRepo repository.CourseRepository, ocrClient *ocr.Client, notificationService NotificationService) AssignmentService {
	return &assignmentService{
		assignmentRepo:      assignmentRepo,
		courseRepo:          courseRepo,
		ocrClient:           ocrClient,
		notificationService: notificationService,
	}
}

func (s *assignmentService) checkTeacherAccess(courseID, userID uuid.UUID) error {
	course, err := s.courseRepo.GetCourseByID(courseID)
	if err != nil {
		return err
	}
	if course.TeacherID != userID {
		return errors.New("forbidden: only the course teacher can perform this action")
	}
	return nil
}

func (s *assignmentService) checkStudentAccess(courseID, userID uuid.UUID) error {
	enrollment, err := s.courseRepo.GetEnrollment(userID, courseID)
	if err != nil {
		return err
	}
	if enrollment == nil {
		return errors.New("forbidden: you must be enrolled in the course to perform this action")
	}
	return nil
}

func (s *assignmentService) CreateAssignment(courseID, teacherID uuid.UUID, req *dto.CreateAssignmentRequest) (*dto.AssignmentResponse, error) {
	if err := s.checkTeacherAccess(courseID, teacherID); err != nil {
		return nil, err
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, req.DueDate)
		if err == nil {
			dueDate = &parsedDate
		}
	}

	assignment := &model.Assignment{
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
		MaxScore:    req.MaxScore,
		DueDate:     dueDate,
	}

	if err := s.assignmentRepo.CreateAssignment(assignment); err != nil {
		return nil, err
	}

	return s.mapAssignmentToDTO(assignment), nil
}

func (s *assignmentService) GetAssignmentsByCourse(courseID, userID uuid.UUID) ([]dto.AssignmentResponse, error) {
	// Let's assume anyone enrolled or the teacher can see the assignments
	isTeacherError := s.checkTeacherAccess(courseID, userID)
	isStudentError := s.checkStudentAccess(courseID, userID)

	// Admin could also see it ideally, but let's just do teacher & student
	if isTeacherError != nil && isStudentError != nil {
		return nil, errors.New("forbidden: access denied to assignments")
	}

	assignments, err := s.assignmentRepo.GetAssignmentsByCourseID(courseID)
	if err != nil {
		return nil, err
	}

	var dtos []dto.AssignmentResponse
	for _, a := range assignments {
		dtos = append(dtos, *s.mapAssignmentToDTO(&a))
	}
	return dtos, nil
}

func (s *assignmentService) GetAssignmentByID(courseID, assignmentID, userID uuid.UUID) (*dto.AssignmentResponse, error) {
	isTeacherError := s.checkTeacherAccess(courseID, userID)
	isStudentError := s.checkStudentAccess(courseID, userID)

	if isTeacherError != nil && isStudentError != nil {
		return nil, errors.New("forbidden: access denied")
	}

	assignment, err := s.assignmentRepo.GetAssignmentByID(assignmentID)
	if err != nil {
		return nil, err
	}

	if assignment.CourseID != courseID {
		return nil, errors.New("assignment does not belong to this course")
	}

	return s.mapAssignmentToDTO(assignment), nil
}

func (s *assignmentService) UpdateAssignment(courseID, assignmentID, teacherID uuid.UUID, req *dto.UpdateAssignmentRequest) (*dto.AssignmentResponse, error) {
	if err := s.checkTeacherAccess(courseID, teacherID); err != nil {
		return nil, err
	}

	assignment, err := s.assignmentRepo.GetAssignmentByID(assignmentID)
	if err != nil {
		return nil, err
	}
	if assignment.CourseID != courseID {
		return nil, errors.New("assignment does not belong to this course")
	}

	if req.Title != nil {
		assignment.Title = *req.Title
	}
	if req.Description != nil {
		assignment.Description = *req.Description
	}
	if req.MaxScore != nil {
		assignment.MaxScore = *req.MaxScore
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			assignment.DueDate = nil
		} else {
			parsedDate, err := time.Parse(time.RFC3339, *req.DueDate)
			if err == nil {
				assignment.DueDate = &parsedDate
			}
		}
	}

	if err := s.assignmentRepo.UpdateAssignment(assignment); err != nil {
		return nil, err
	}

	return s.mapAssignmentToDTO(assignment), nil
}

func (s *assignmentService) DeleteAssignment(courseID, assignmentID, teacherID uuid.UUID) error {
	if err := s.checkTeacherAccess(courseID, teacherID); err != nil {
		return err
	}

	assignment, err := s.assignmentRepo.GetAssignmentByID(assignmentID)
	if err != nil {
		return err
	}
	if assignment.CourseID != courseID {
		return errors.New("assignment does not belong to this course")
	}

	return s.assignmentRepo.DeleteAssignment(assignmentID)
}

// mapAssignmentToDTO is a helper function
func (s *assignmentService) mapAssignmentToDTO(a *model.Assignment) *dto.AssignmentResponse {
	res := &dto.AssignmentResponse{
		ID:          a.ID,
		CourseID:    a.CourseID,
		Title:       a.Title,
		Description: a.Description,
		MaxScore:    a.MaxScore,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
	}
	if a.DueDate != nil {
		dateStr := a.DueDate.Format(time.RFC3339)
		res.DueDate = &dateStr
	}
	return res
}

func (s *assignmentService) SubmitAssignment(courseID, assignmentID, studentID uuid.UUID, req *dto.SubmitAssignmentRequest) (*dto.AssignmentSubmissionResponse, error) {
	if err := s.checkStudentAccess(courseID, studentID); err != nil {
		return nil, err
	}

	assignment, err := s.assignmentRepo.GetAssignmentByID(assignmentID)
	if err != nil {
		return nil, err
	}
	if assignment.CourseID != courseID {
		return nil, errors.New("assignment does not belong to this course")
	}

	// Check if already submitted
	existing, err := s.assignmentRepo.GetSubmissionByUserAndAssignment(studentID, assignmentID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("assignment already submitted. Please delete or update (if supported) instead")
	}

	// Extract the actual file path from the URL. Assuming the file is mounted/stored locally.
	// We parse the URL to get just the path part, e.g. "/uploads/images/abc.jpg"
	localFilePath := req.FileURL
	if parsedURL, err := url.Parse(req.FileURL); err == nil {
		localFilePath = "." + parsedURL.Path // e.g. "./uploads/images/abc.jpg"
	}

	extractedText := ""
	if s.ocrClient != nil && req.FileURL != "" {
		// As this might take a few seconds, you could run it in a goroutine if you wanted a fire-and-forget.
		// However, for immediate feedback to the teacher, we'll run it synchronously here if it's not too large.
		text, err := s.ocrClient.ExtractText(localFilePath)
		if err == nil {
			extractedText = text
		} else {
			extractedText = "Error during OCR extraction: " + err.Error()
		}
	} else {
		extractedText = "OCR Extraction pending or no client configured..."
	}

	submission := &model.AssignmentSubmission{
		AssignmentID:  assignmentID,
		UserID:        studentID,
		FileURL:       req.FileURL,
		ExtractedText: extractedText,
	}

	if err := s.assignmentRepo.CreateSubmission(submission); err != nil {
		return nil, err
	}

	// Notify Teacher
	course, err := s.courseRepo.GetCourseByID(courseID)
	if err == nil && course != nil {
		s.notificationService.CreateNotification(
			course.TeacherID,
			"New Assignment Submission",
			"A student has submitted an assignment for "+assignment.Title,
			"assignment_submitted",
		)
	}

	return s.mapSubmissionToDTO(submission), nil
}

func (s *assignmentService) GetSubmissions(courseID, assignmentID, teacherID uuid.UUID) ([]dto.AssignmentSubmissionResponse, error) {
	if err := s.checkTeacherAccess(courseID, teacherID); err != nil {
		return nil, err
	}

	submissions, err := s.assignmentRepo.GetSubmissionsByAssignmentID(assignmentID)
	if err != nil {
		return nil, err
	}

	var dtos []dto.AssignmentSubmissionResponse
	for _, sub := range submissions {
		dtos = append(dtos, *s.mapSubmissionToDTO(&sub))
	}
	return dtos, nil
}

func (s *assignmentService) GradeSubmission(courseID, assignmentID, submissionID, teacherID uuid.UUID, req *dto.GradeSubmissionRequest) (*dto.AssignmentSubmissionResponse, error) {
	if err := s.checkTeacherAccess(courseID, teacherID); err != nil {
		return nil, err
	}

	submission, err := s.assignmentRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}
	if submission.AssignmentID != assignmentID {
		return nil, errors.New("submission does not belong to this assignment")
	}

	now := time.Now()
	submission.Score = &req.Score
	submission.Feedback = req.Feedback
	submission.GradedAt = &now

	if err := s.assignmentRepo.UpdateSubmission(submission); err != nil {
		return nil, err
	}

	// Notify Student
	assignment, err := s.assignmentRepo.GetAssignmentByID(assignmentID)
	if err == nil && assignment != nil {
		s.notificationService.CreateNotification(
			submission.UserID,
			"Assignment Graded",
			"Your submission for "+assignment.Title+" has been graded.",
			"assignment_graded",
		)
	}

	return s.mapSubmissionToDTO(submission), nil
}

func (s *assignmentService) GetStudentSubmission(courseID, assignmentID, studentID uuid.UUID) (*dto.AssignmentSubmissionResponse, error) {
	if err := s.checkStudentAccess(courseID, studentID); err != nil {
		return nil, err
	}

	submission, err := s.assignmentRepo.GetSubmissionByUserAndAssignment(studentID, assignmentID)
	if err != nil {
		return nil, err
	}
	if submission == nil {
		// Just return an empty response or custom status indicating no submission yet
		// We'll return nil, nil to signify not found but no error to 404 cleanly
		return nil, nil
	}

	return s.mapSubmissionToDTO(submission), nil
}

func (s *assignmentService) mapSubmissionToDTO(sub *model.AssignmentSubmission) *dto.AssignmentSubmissionResponse {
	res := &dto.AssignmentSubmissionResponse{
		ID:            sub.ID,
		AssignmentID:  sub.AssignmentID,
		UserID:        sub.UserID,
		FileURL:       sub.FileURL,
		ExtractedText: sub.ExtractedText,
		Score:         sub.Score,
		Feedback:      sub.Feedback,
		SubmittedAt:   sub.SubmittedAt.Format(time.RFC3339),
	}
	if sub.User.Profile.FirstName != "" { // Assuming preloaded
		res.UserName = sub.User.Profile.FirstName + " " + sub.User.Profile.LastName
	}
	if sub.GradedAt != nil {
		dateStr := sub.GradedAt.Format(time.RFC3339)
		res.GradedAt = &dateStr
	}
	return res
}
