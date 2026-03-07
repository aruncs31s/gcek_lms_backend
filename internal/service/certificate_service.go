package service

import (
	"errors"
	"time"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/aruncs/esdc-lms/pkg/certgen"
	"github.com/google/uuid"
)

type CertificateService interface {
	GenerateCertificate(
		req *dto.GenerateCertificateRequest,
		baseURL string,
	) (*dto.CertificateResponse, error)
}

type certificateService struct {
	certRepo     repository.CertificateRepository
	userRepo     repository.UserRepository
	courseRepo   repository.CourseRepository
	orchestrator *certgen.Orchestrator
}

func NewCertificateService(
	cr repository.CertificateRepository,
	ur repository.UserRepository,
	cor repository.CourseRepository,
	orc *certgen.Orchestrator,
) CertificateService {
	return &certificateService{
		certRepo:     cr,
		userRepo:     ur,
		courseRepo:   cor,
		orchestrator: orc,
	}
}

func (s *certificateService) GenerateCertificate(req *dto.GenerateCertificateRequest, baseURL string) (*dto.CertificateResponse, error) {
	// Parse IDs
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		return nil, errors.New("invalid course ID")
	}

	// Fetch required User and Course information
	// In production we would join Course -> Teacher -> Profile to get TeacherName
	// For MVP, we'll dummy check if user is completing (in reality, query enrollments)
	course, err := s.courseRepo.GetCourseByID(courseID)
	if err != nil || course == nil {
		return nil, errors.New("course not found")
	}

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	teacherName := "ESDC LMS Instructor"
	if course.TeacherID != uuid.Nil {
		teacher, err := s.userRepo.GetUserByID(course.TeacherID)
		if err == nil && teacher != nil {
			teacherName = teacher.Profile.FirstName + " " + teacher.Profile.LastName
		}
	}

	docModel := certgen.DocumentModel{
		StudentName: user.Profile.FirstName + " " + user.Profile.LastName,
		CourseName:  course.Title,
		TeacherName: teacherName,
		DateIssued:  time.Now().Format("Jan 02, 2006"),
	}

	// Orchestrator executes Chromeless print
	fileName, err := s.orchestrator.GeneratePDF(docModel)
	if err != nil {
		return nil, err
	}

	fileURL := baseURL + "/uploads/certificates/" + fileName

	// Save to DB
	cert := &model.Certificate{
		UserID:   userID,
		CourseID: courseID,
		FileURL:  fileURL,
	}

	if err := s.certRepo.SaveCertificate(cert); err != nil {
		return nil, err
	}

	return &dto.CertificateResponse{
		ID:       cert.ID.String(),
		UserID:   cert.UserID.String(),
		CourseID: cert.CourseID.String(),
		FileURL:  cert.FileURL,
		IssuedAt: cert.IssuedAt,
	}, nil
}
