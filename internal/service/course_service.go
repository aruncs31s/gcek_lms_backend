package service

import (
	"context"
	"errors"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/google/uuid"
)

type CourseService interface {
	CreateCourse(
		teacherID uuid.UUID,
		req *dto.CreateCourseRequest,
	) (*dto.CourseResponse, error)
	GetCourseByID(
		id uuid.UUID,
		userID uuid.UUID,
	) (*dto.CourseResponse, error)
	GetAllCourses(
		userID uuid.UUID,
		query string,
		courseType string,
		format string,
		status string,
		teacherID string,
	) ([]dto.CourseResponse, error)
	UpdateCourse(
		id uuid.UUID,
		teacherID uuid.UUID,
		req *dto.UpdateCourseRequest,
	) (*dto.CourseResponse, error)
	DeleteCourse(
		id uuid.UUID,
		teacherID uuid.UUID,
	) error

	CreateModule(
		courseID uuid.UUID,
		teacherID uuid.UUID,
		req *dto.CreateModuleRequest,
	) (*dto.ModuleResponse, error)
	UpdateModule(
		courseID uuid.UUID,
		moduleID uuid.UUID,
		teacherID uuid.UUID,
		req *dto.UpdateModuleRequest,
	) (*dto.ModuleResponse, error)
	DeleteModule(
		id uuid.UUID,
		teacherID uuid.UUID,
	) error
	ReorderModules(
		courseID uuid.UUID,
		teacherID uuid.UUID,
		req *dto.ReorderModulesRequest,
	) error

	EnrollCourse(
		courseID uuid.UUID,
		userID uuid.UUID,
	) error
	GetEnrollmentStatus(
		courseID uuid.UUID,
		userID uuid.UUID,
	) (*model.Enrollment, error)

	CompleteModule(
		courseID uuid.UUID,
		moduleID uuid.UUID,
		userID uuid.UUID,
	) error

	LikeCourse(
		courseID uuid.UUID,
		userID uuid.UUID,
	) error
	UnlikeCourse(
		courseID uuid.UUID,
		userID uuid.UUID,
	) error
	GetTrendingCourses(
		limit int,
		userID uuid.UUID,
	) ([]dto.CourseResponse, error)

	AddReview(
		courseID uuid.UUID,
		userID uuid.UUID,
		req *dto.CreateCourseReviewRequest,
	) (*dto.CourseReviewResponse, error)
	GetReviews(
		courseID uuid.UUID,
	) ([]dto.CourseReviewResponse, error)

	SearchCourses(
		query string,
		courseType string,
		format string,
		status string,
		limit int,
		offset int,
	) ([]dto.CourseSearchResponse, error)
	GetEnrolledUsers(
		ctx context.Context,
		courseID uuid.UUID,
	) ([]dto.UserResponse, error)
	GetEnrolledUsersCount(
		ctx context.Context,
		courseID uuid.UUID,
	) (int64, error)
}

type courseService struct {
	courseRepo repository.CourseRepository
	ur         repository.UserRepository
}

func NewCourseService(
	repo repository.CourseRepository,
	ur repository.UserRepository,
) CourseService {
	return &courseService{
		courseRepo: repo,
		ur:         ur,
	}
}

func (s *courseService) CreateCourse(teacherID uuid.UUID, req *dto.CreateCourseRequest) (*dto.CourseResponse, error) {
	course := &model.Course{
		TeacherID:              teacherID,
		Title:                  req.Title,
		Description:            req.Description,
		ThumbnailURL:           req.ThumbnailURL,
		Price:                  req.Price,
		Type:                   req.Type,
		Format:                 req.Format,
		Status:                 req.Status,
		Duration:               req.Duration,
		IsCertificateAvailable: req.IsCertificateAvailable,
		StartDate:              req.StartDate,
	}

	if course.Type == "" {
		course.Type = "paid"
	}
	if course.Format == "" {
		course.Format = "course"
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

func (s *courseService) GetAllCourses(
	userID uuid.UUID,
	query string,
	courseType string,
	format string,
	status string,
	teacherID string,
) ([]dto.CourseResponse, error) {
	courses, err := s.courseRepo.GetAllCourses(query, courseType, format, status, teacherID)
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
	if req.Format != nil {
		course.Format = *req.Format
	}
	if req.Status != nil {
		course.Status = *req.Status
	}
	if req.Duration != nil {
		course.Duration = *req.Duration
	}
	if req.IsCertificateAvailable != nil {
		course.IsCertificateAvailable = *req.IsCertificateAvailable
	}
	if req.StartDate != nil {
		course.StartDate = req.StartDate
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
	module, err := s.courseRepo.GetModuleByID(id)
	if err != nil {
		return err
	}
	if module == nil {
		return errors.New("module not found")
	}

	course, err := s.courseRepo.GetCourseByID(module.CourseID)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.New("associated course not found")
	}

	if course.TeacherID != teacherID {
		return errors.New("forbidden: only the course creator can delete modules")
	}

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
		ID:                     course.ID.String(),
		TeacherID:              course.TeacherID.String(),
		TeacherName:            course.Teacher.Profile.FirstName + " " + course.Teacher.Profile.LastName,
		TeacherAvatar:          course.Teacher.Profile.AvatarURL,
		TeacherBio:             course.Teacher.Profile.Bio,
		Title:                  course.Title,
		Description:            course.Description,
		ThumbnailURL:           course.ThumbnailURL,
		Price:                  course.Price,
		Type:                   course.Type,
		Format:                 course.Format,
		Status:                 course.Status,
		Duration:               course.Duration,
		IsCertificateAvailable: course.IsCertificateAvailable,
		StartDate:              course.StartDate,
		CreatedAt:              course.CreatedAt,
		StudentCount:           len(course.Enrollments),
		LikesCount:             likesCount,
		IsLiked:                isLiked,
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

		// calculate basic progress (videos only)
		completedCount := 0
		videoCount := 0
		for _, m := range course.Modules {
			if m.Type == "video" {
				videoCount++
				if completedModules[m.ID] {
					completedCount++
				}
			}
		}

		if videoCount > 0 {
			resp.Progress = float64(completedCount) / float64(videoCount) * 100
		}

		resp.Modules = modResp
	}

	return resp
}

func (s *courseService) AddReview(courseID uuid.UUID, userID uuid.UUID, req *dto.CreateCourseReviewRequest) (*dto.CourseReviewResponse, error) {
	// Let's assume anyone who is enrolled or has completed can review, or for MVP anyone logged in can review
	// We'll just allow it if course exists.
	course, err := s.courseRepo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}

	review := &model.CourseReview{
		CourseID: courseID,
		UserID:   userID,
		Rating:   req.Rating,
		Comment:  req.Comment,
	}

	if err := s.courseRepo.CreateReview(review); err != nil {
		return nil, err
	}

	// We don't have user info loaded in this object immediately without a refetch,
	// but we can return basic for now or refetch the reviews list.
	// For simplicity, we just return the basics. The frontend will likely reload reviews.
	return &dto.CourseReviewResponse{
		ID:        review.ID.String(),
		CourseID:  review.CourseID.String(),
		UserID:    review.UserID.String(),
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt,
	}, nil
}

func (s *courseService) GetReviews(courseID uuid.UUID) ([]dto.CourseReviewResponse, error) {
	reviews, err := s.courseRepo.GetReviewsByCourseID(courseID)
	if err != nil {
		return nil, err
	}

	var res []dto.CourseReviewResponse
	for _, r := range reviews {
		res = append(res, dto.CourseReviewResponse{
			ID:        r.ID.String(),
			CourseID:  r.CourseID.String(),
			UserID:    r.UserID.String(),
			UserName:  r.User.Profile.FirstName + " " + r.User.Profile.LastName,
			AvatarURL: r.User.Profile.AvatarURL,
			Rating:    r.Rating,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt,
		})
	}

	if res == nil {
		res = []dto.CourseReviewResponse{}
	}

	return res, nil
}

func (s *courseService) SearchCourses(query string, courseType string, format string, status string, limit int, offset int) ([]dto.CourseSearchResponse, error) {
	courses, err := s.courseRepo.SearchCourses(query, courseType, format, status, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []dto.CourseSearchResponse
	for _, c := range courses {
		responses = append(responses, dto.CourseSearchResponse{
			ID:           c.ID.String(),
			Title:        c.Title,
			ThumbnailURL: c.ThumbnailURL,
		})
	}
	if responses == nil {
		responses = []dto.CourseSearchResponse{}
	}
	return responses, nil
}

func (s *courseService) GetEnrolledUsers(ctx context.Context, courseID uuid.UUID) ([]dto.UserResponse, error) {
	users, err := s.ur.GetUsersByCourse(ctx, courseID)
	if err != nil {
		return []dto.UserResponse{}, err
	}
	m := NewUserMapper()
	var res []dto.UserResponse
	for _, u := range users {
		res = append(res, *m.MapToDTO(&u))
	}
	if res == nil {
		res = []dto.UserResponse{}
	}
	return res, nil
}

func (s *courseService) GetEnrolledUsersCount(ctx context.Context, courseID uuid.UUID) (int64, error) {
	return s.ur.GetUserCountByCourse(ctx, courseID)
}
