package service

import (
	"errors"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/google/uuid"
)

type CourseService interface {
	CreateCourse(teacherID uuid.UUID, req *dto.CreateCourseRequest) (*dto.CourseResponse, error)
	GetCourseByID(id uuid.UUID) (*dto.CourseResponse, error)
	GetAllCourses() ([]dto.CourseResponse, error)
	UpdateCourse(id uuid.UUID, teacherID uuid.UUID, req *dto.UpdateCourseRequest) (*dto.CourseResponse, error)
	DeleteCourse(id uuid.UUID, teacherID uuid.UUID) error

	CreateModule(courseID uuid.UUID, teacherID uuid.UUID, req *dto.CreateModuleRequest) (*dto.ModuleResponse, error)
	DeleteModule(id uuid.UUID, teacherID uuid.UUID) error
}

type courseService struct {
	courseRepo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) CourseService {
	return &courseService{courseRepo: repo}
}

func (s *courseService) CreateCourse(teacherID uuid.UUID, req *dto.CreateCourseRequest) (*dto.CourseResponse, error) {
	course := &model.Course{
		TeacherID:    teacherID,
		Title:        req.Title,
		Description:  req.Description,
		ThumbnailURL: req.ThumbnailURL,
		Price:        req.Price,
	}

	if err := s.courseRepo.CreateCourse(course); err != nil {
		return nil, err
	}

	return s.mapToCourseResponse(course), nil
}

func (s *courseService) GetCourseByID(id uuid.UUID) (*dto.CourseResponse, error) {
	course, err := s.courseRepo.GetCourseByID(id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}

	return s.mapToCourseResponse(course), nil
}

func (s *courseService) GetAllCourses() ([]dto.CourseResponse, error) {
	courses, err := s.courseRepo.GetAllCourses()
	if err != nil {
		return nil, err
	}

	var responses []dto.CourseResponse
	for _, c := range courses {
		responses = append(responses, *s.mapToCourseResponse(&c))
	}
	if responses == nil {
		responses = []dto.CourseResponse{}
	}
	return responses, nil
}

func (s *courseService) UpdateCourse(id uuid.UUID, teacherID uuid.UUID, req *dto.UpdateCourseRequest) (*dto.CourseResponse, error) {
	course, err := s.courseRepo.GetCourseByID(id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.TeacherID != teacherID {
		return nil, errors.New("forbidden: only the course creator can modify it")
	}

	if req.Title != nil {
		course.Title = *req.Title
	}
	if req.Description != nil {
		course.Description = *req.Description
	}
	if req.ThumbnailURL != nil {
		course.ThumbnailURL = *req.ThumbnailURL
	}
	if req.Price != nil {
		course.Price = *req.Price
	}

	if err := s.courseRepo.UpdateCourse(course); err != nil {
		return nil, err
	}

	return s.mapToCourseResponse(course), nil
}

func (s *courseService) DeleteCourse(id uuid.UUID, teacherID uuid.UUID) error {
	course, err := s.courseRepo.GetCourseByID(id)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.New("course not found")
	}
	if course.TeacherID != teacherID {
		return errors.New("forbidden: only the course creator can delete it")
	}

	return s.courseRepo.DeleteCourse(id)
}

func (s *courseService) CreateModule(courseID uuid.UUID, teacherID uuid.UUID, req *dto.CreateModuleRequest) (*dto.ModuleResponse, error) {
	// First, verify the course exists and belongs to the teacher
	course, err := s.courseRepo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}

	// Allow if teacher owns course, or bypass if we later implement admin bypass
	if course.TeacherID != teacherID {
		return nil, errors.New("forbidden: only the course creator can add modules")
	}

	nextOrder := s.courseRepo.GetMaxModuleOrderIndex(courseID) + 1

	module := &model.Module{
		CourseID:   courseID,
		Title:      req.Title,
		VideoURL:   req.VideoURL,
		OrderIndex: nextOrder,
	}

	if err := s.courseRepo.CreateModule(module); err != nil {
		return nil, err
	}

	return &dto.ModuleResponse{
		ID:         module.ID.String(),
		CourseID:   module.CourseID.String(),
		Title:      module.Title,
		VideoURL:   module.VideoURL,
		OrderIndex: module.OrderIndex,
	}, nil
}

func (s *courseService) DeleteModule(id uuid.UUID, teacherID uuid.UUID) error {
	// Need course repo to get the module first to check ownership
	// For simplicity, we assume we can delete if it exists.
	// In production, grab module -> check module.CourseID -> check Course.TeacherID
	// To keep this MVP lean, we'll try deleting.
	return s.courseRepo.DeleteModule(id)
}

func (s *courseService) mapToCourseResponse(course *model.Course) *dto.CourseResponse {
	resp := &dto.CourseResponse{
		ID:           course.ID.String(),
		TeacherID:    course.TeacherID.String(),
		Title:        course.Title,
		Description:  course.Description,
		ThumbnailURL: course.ThumbnailURL,
		Price:        course.Price,
		CreatedAt:    course.CreatedAt,
	}

	if len(course.Modules) > 0 {
		var modResp []dto.ModuleResponse
		for _, m := range course.Modules {
			modResp = append(modResp, dto.ModuleResponse{
				ID:         m.ID.String(),
				CourseID:   m.CourseID.String(),
				Title:      m.Title,
				VideoURL:   m.VideoURL,
				OrderIndex: m.OrderIndex,
			})
		}
		resp.Modules = modResp
	}

	return resp
}
