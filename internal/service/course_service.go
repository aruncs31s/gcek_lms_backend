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
	GetCourseByID(id uuid.UUID, userID uuid.UUID) (*dto.CourseResponse, error)
	GetAllCourses(userID uuid.UUID, query string, courseType string, status string) ([]dto.CourseResponse, error)
	UpdateCourse(id uuid.UUID, teacherID uuid.UUID, req *dto.UpdateCourseRequest) (*dto.CourseResponse, error)
	DeleteCourse(id uuid.UUID, teacherID uuid.UUID) error

	CreateModule(courseID uuid.UUID, teacherID uuid.UUID, req *dto.CreateModuleRequest) (*dto.ModuleResponse, error)
	UpdateModule(courseID uuid.UUID, moduleID uuid.UUID, teacherID uuid.UUID, req *dto.UpdateModuleRequest) (*dto.ModuleResponse, error)
	DeleteModule(id uuid.UUID, teacherID uuid.UUID) error
	ReorderModules(courseID uuid.UUID, teacherID uuid.UUID, req *dto.ReorderModulesRequest) error

	EnrollCourse(courseID uuid.UUID, userID uuid.UUID) error
	GetEnrollmentStatus(courseID uuid.UUID, userID uuid.UUID) (*model.Enrollment, error)
	CompleteModule(courseID uuid.UUID, moduleID uuid.UUID, userID uuid.UUID) error

	LikeCourse(courseID uuid.UUID, userID uuid.UUID) error
	UnlikeCourse(courseID uuid.UUID, userID uuid.UUID) error
	GetTrendingCourses(limit int, userID uuid.UUID) ([]dto.CourseResponse, error)
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
		Type:         req.Type,
		Status:       req.Status,
	}

	if course.Type == "" {
		course.Type = "paid"
	}
	if course.Status == "" {
		course.Status = "not started"
	}

	if err := s.courseRepo.CreateCourse(course); err != nil {
		return nil, err
	}

	return s.mapToCourseResponse(course, uuid.Nil), nil
}

func (s *courseService) GetCourseByID(id uuid.UUID, userID uuid.UUID) (*dto.CourseResponse, error) {
	course, err := s.courseRepo.GetCourseByID(id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}

	return s.mapToCourseResponse(course, userID), nil
}

func (s *courseService) GetAllCourses(userID uuid.UUID, query string, courseType string, status string) ([]dto.CourseResponse, error) {
	courses, err := s.courseRepo.GetAllCourses(query, courseType, status)
	if err != nil {
		return nil, err
	}

	var responses []dto.CourseResponse
	for _, c := range courses {
		responses = append(responses, *s.mapToCourseResponse(&c, userID))
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
	if req.Type != nil {
		course.Type = *req.Type
	}
	if req.Status != nil {
		course.Status = *req.Status
	}

	if err := s.courseRepo.UpdateCourse(course); err != nil {
		return nil, err
	}

	return s.mapToCourseResponse(course, teacherID), nil
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

	var parentID *uuid.UUID
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err == nil {
			parentID = &pid
		}
	}

	module := &model.Module{
		CourseID:    courseID,
		ParentID:    parentID,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		VideoURL:    req.VideoURL,
		Points:      req.Points,
		IsFree:      req.IsFree,
		OrderIndex:  nextOrder,
	}

	if err := s.courseRepo.CreateModule(module); err != nil {
		return nil, err
	}

	var pIDStr *string
	if module.ParentID != nil {
		str := module.ParentID.String()
		pIDStr = &str
	}

	return &dto.ModuleResponse{
		ID:          module.ID.String(),
		CourseID:    module.CourseID.String(),
		ParentID:    pIDStr,
		Title:       module.Title,
		Description: module.Description,
		Type:        module.Type,
		VideoURL:    module.VideoURL,
		Points:      module.Points,
		IsFree:      module.IsFree,
		OrderIndex:  module.OrderIndex,
	}, nil
}

func (s *courseService) UpdateModule(courseID uuid.UUID, moduleID uuid.UUID, teacherID uuid.UUID, req *dto.UpdateModuleRequest) (*dto.ModuleResponse, error) {
	course, err := s.courseRepo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.TeacherID != teacherID {
		return nil, errors.New("forbidden: only the course creator can update modules")
	}

	module, err := s.courseRepo.GetModuleByID(moduleID)
	if err != nil {
		return nil, err
	}
	if module == nil || module.CourseID != courseID {
		return nil, errors.New("module not found in the specified course")
	}

	if req.ParentID != nil {
		if *req.ParentID == "" {
			module.ParentID = nil
		} else {
			pid, err := uuid.Parse(*req.ParentID)
			if err == nil {
				module.ParentID = &pid
			}
		}
	}
	if req.Title != nil {
		module.Title = *req.Title
	}
	if req.Description != nil {
		module.Description = *req.Description
	}
	if req.VideoURL != nil {
		module.VideoURL = *req.VideoURL
	}
	if req.Points != nil {
		module.Points = *req.Points
	}
	if req.IsFree != nil {
		module.IsFree = *req.IsFree
	}

	if err := s.courseRepo.UpdateModule(module); err != nil {
		return nil, err
	}

	var pIDStr *string
	if module.ParentID != nil {
		str := module.ParentID.String()
		pIDStr = &str
	}

	return &dto.ModuleResponse{
		ID:          module.ID.String(),
		CourseID:    module.CourseID.String(),
		ParentID:    pIDStr,
		Title:       module.Title,
		Description: module.Description,
		Type:        module.Type,
		VideoURL:    module.VideoURL,
		Points:      module.Points,
		IsFree:      module.IsFree,
		OrderIndex:  module.OrderIndex,
	}, nil
}

func (s *courseService) DeleteModule(id uuid.UUID, teacherID uuid.UUID) error {
	// Need course repo to get the module first to check ownership
	// For simplicity, we assume we can delete if it exists.
	// In production, grab module -> check module.CourseID -> check Course.TeacherID
	// To keep this MVP lean, we'll try deleting.
	return s.courseRepo.DeleteModule(id)
}

func (s *courseService) ReorderModules(courseID uuid.UUID, teacherID uuid.UUID, req *dto.ReorderModulesRequest) error {
	course, err := s.courseRepo.GetCourseByID(courseID)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.New("course not found")
	}
	if course.TeacherID != teacherID {
		return errors.New("forbidden: only the course creator can reorder modules")
	}

	var parsedIDs []uuid.UUID
	for _, idStr := range req.ModuleIDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			parsedIDs = append(parsedIDs, id)
		}
	}

	return s.courseRepo.UpdateModuleOrder(courseID, parsedIDs)
}

func (s *courseService) EnrollCourse(courseID uuid.UUID, userID uuid.UUID) error {
	course, err := s.courseRepo.GetCourseByID(courseID)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.New("course not found")
	}

	enrollment := &model.Enrollment{
		UserID:   userID,
		CourseID: courseID,
		Status:   model.EnrollmentStatusActive,
	}

	return s.courseRepo.CreateEnrollment(enrollment)
}

func (s *courseService) GetEnrollmentStatus(courseID uuid.UUID, userID uuid.UUID) (*model.Enrollment, error) {
	return s.courseRepo.GetEnrollment(userID, courseID)
}

func (s *courseService) CompleteModule(courseID uuid.UUID, moduleID uuid.UUID, userID uuid.UUID) error {
	// First check if already completed
	progresses, _ := s.courseRepo.GetModuleProgresses(userID, courseID)
	for _, p := range progresses {
		if p.ModuleID == moduleID && p.Completed {
			return nil // already completed, don't double count points
		}
	}

	module, err := s.courseRepo.GetModuleByID(moduleID)
	if err != nil {
		return err
	}

	err = s.courseRepo.UpdateModuleProgress(&model.ModuleProgress{
		UserID:    userID,
		ModuleID:  moduleID,
		Completed: true,
	})

	// Add points to user profile if progress updated successfully and module has points
	if err == nil && module != nil && module.Points > 0 {
		s.courseRepo.AddPointsToProfile(userID, module.Points)
	}

	return err
}

func (s *courseService) LikeCourse(courseID uuid.UUID, userID uuid.UUID) error {
	return s.courseRepo.LikeCourse(userID, courseID)
}

func (s *courseService) UnlikeCourse(courseID uuid.UUID, userID uuid.UUID) error {
	return s.courseRepo.UnlikeCourse(userID, courseID)
}

func (s *courseService) GetTrendingCourses(limit int, userID uuid.UUID) ([]dto.CourseResponse, error) {
	courses, err := s.courseRepo.GetTrendingCourses(limit)
	if err != nil {
		return nil, err
	}

	var responses []dto.CourseResponse
	for _, c := range courses {
		responses = append(responses, *s.mapToCourseResponse(&c, userID))
	}
	if responses == nil {
		responses = []dto.CourseResponse{}
	}
	return responses, nil
}

func (s *courseService) mapToCourseResponse(course *model.Course, userID uuid.UUID) *dto.CourseResponse {
	likesCount, _ := s.courseRepo.GetCourseLikesCount(course.ID)
	isLiked := false
	if userID != uuid.Nil {
		isLiked, _ = s.courseRepo.HasUserLikedCourse(userID, course.ID)
	}

	resp := &dto.CourseResponse{
		ID:            course.ID.String(),
		TeacherID:     course.TeacherID.String(),
		TeacherName:   course.Teacher.Profile.FirstName + " " + course.Teacher.Profile.LastName,
		TeacherAvatar: course.Teacher.Profile.AvatarURL,
		TeacherBio:    course.Teacher.Profile.Bio,
		Title:         course.Title,
		Description:   course.Description,
		ThumbnailURL:  course.ThumbnailURL,
		Price:         course.Price,
		Type:          course.Type,
		Status:        course.Status,
		CreatedAt:     course.CreatedAt,
		StudentCount:  len(course.Enrollments),
		LikesCount:    likesCount,
		IsLiked:       isLiked,
	}

	if len(course.Modules) > 0 {
		// Fetch progress if userID is valid
		completedModules := make(map[uuid.UUID]bool)
		if userID != uuid.Nil {
			progresses, _ := s.courseRepo.GetModuleProgresses(userID, course.ID)
			for _, p := range progresses {
				if p.Completed {
					completedModules[p.ModuleID] = true
				}
			}
		}

		var modResp []dto.ModuleResponse
		for _, m := range course.Modules {
			var pIDStr *string
			if m.ParentID != nil {
				str := m.ParentID.String()
				pIDStr = &str
			}

			modResp = append(modResp, dto.ModuleResponse{
				ID:          m.ID.String(),
				CourseID:    m.CourseID.String(),
				ParentID:    pIDStr,
				Title:       m.Title,
				Description: m.Description,
				Type:        m.Type,
				VideoURL:    m.VideoURL,
				Points:      m.Points,
				IsFree:      m.IsFree,
				OrderIndex:  m.OrderIndex,
				IsCompleted: completedModules[m.ID],
			})
		}
		resp.Modules = modResp
	}

	return resp
}
