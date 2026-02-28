package repository

import (
	"errors"

	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CourseRepository interface {
	CreateCourse(course *model.Course) error
	GetCourseByID(id uuid.UUID) (*model.Course, error)
	GetAllCourses(query string, courseType string, status string) ([]model.Course, error)
	UpdateCourse(course *model.Course) error
	DeleteCourse(id uuid.UUID) error

	CreateModule(module *model.Module) error
	GetModuleByID(id uuid.UUID) (*model.Module, error)
	GetModulesByCourseID(courseID uuid.UUID) ([]model.Module, error)
	GetMaxModuleOrderIndex(courseID uuid.UUID) int
	UpdateModule(module *model.Module) error
	DeleteModule(id uuid.UUID) error
	UpdateModuleOrder(courseID uuid.UUID, orderedIDs []uuid.UUID) error

	CreateEnrollment(enrollment *model.Enrollment) error
	GetEnrollment(userID uuid.UUID, courseID uuid.UUID) (*model.Enrollment, error)

	UpdateModuleProgress(progress *model.ModuleProgress) error
	GetModuleProgresses(userID uuid.UUID, courseID uuid.UUID) ([]model.ModuleProgress, error)

	// Course Likes and Trending
	LikeCourse(userID uuid.UUID, courseID uuid.UUID) error
	UnlikeCourse(userID uuid.UUID, courseID uuid.UUID) error
	HasUserLikedCourse(userID uuid.UUID, courseID uuid.UUID) (bool, error)
	GetCourseLikesCount(courseID uuid.UUID) (int64, error)
	GetTrendingCourses(limit int) ([]model.Course, error)
	AddPointsToProfile(userID uuid.UUID, points int) error

	CreateReview(review *model.CourseReview) error
	GetReviewsByCourseID(courseID uuid.UUID) ([]model.CourseReview, error)
}

type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) CreateCourse(course *model.Course) error {
	return r.db.Create(course).Error
}

func (r *courseRepository) GetCourseByID(id uuid.UUID) (*model.Course, error) {
	var course model.Course
	err := r.db.Preload("Teacher.Profile").Preload("Modules").Preload("Enrollments").First(&course, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &course, nil
}

func (r *courseRepository) GetAllCourses(query string, courseType string, status string) ([]model.Course, error) {
	var courses []model.Course
	db := r.db.Preload("Teacher.Profile").Preload("Enrollments")

	if query != "" {
		db = db.Where("LOWER(title) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", "%"+query+"%", "%"+query+"%")
	}
	if courseType != "" {
		db = db.Where("type = ?", courseType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}

	err := db.Order("created_at desc").Find(&courses).Error
	return courses, err
}

func (r *courseRepository) UpdateCourse(course *model.Course) error {
	return r.db.Save(course).Error
}

func (r *courseRepository) DeleteCourse(id uuid.UUID) error {
	return r.db.Delete(&model.Course{}, "id = ?", id).Error
}

func (r *courseRepository) CreateModule(module *model.Module) error {
	return r.db.Create(module).Error
}

func (r *courseRepository) GetModulesByCourseID(courseID uuid.UUID) ([]model.Module, error) {
	var modules []model.Module
	err := r.db.Where("course_id = ?", courseID).Order("order_index asc").Find(&modules).Error
	return modules, err
}

func (r *courseRepository) GetMaxModuleOrderIndex(courseID uuid.UUID) int {
	var maxOrder int
	r.db.Model(&model.Module{}).Where("course_id = ?", courseID).Select("COALESCE(MAX(order_index), 0)").Scan(&maxOrder)
	return maxOrder
}

func (r *courseRepository) GetModuleByID(id uuid.UUID) (*model.Module, error) {
	var module model.Module
	err := r.db.First(&module, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &module, nil
}

func (r *courseRepository) UpdateModule(module *model.Module) error {
	return r.db.Save(module).Error
}

func (r *courseRepository) DeleteModule(id uuid.UUID) error {
	return r.db.Delete(&model.Module{}, "id = ?", id).Error
}

func (r *courseRepository) UpdateModuleOrder(courseID uuid.UUID, orderedIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range orderedIDs {
			if err := tx.Model(&model.Module{}).Where("id = ? AND course_id = ?", id, courseID).Update("order_index", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *courseRepository) CreateEnrollment(enrollment *model.Enrollment) error {
	return r.db.Create(enrollment).Error
}

func (r *courseRepository) GetEnrollment(userID uuid.UUID, courseID uuid.UUID) (*model.Enrollment, error) {
	var enrollment model.Enrollment
	err := r.db.First(&enrollment, "user_id = ? AND course_id = ?", userID, courseID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &enrollment, nil
}

func (r *courseRepository) UpdateModuleProgress(progress *model.ModuleProgress) error {
	return r.db.Save(progress).Error
}

func (r *courseRepository) GetModuleProgresses(userID uuid.UUID, courseID uuid.UUID) ([]model.ModuleProgress, error) {
	var progresses []model.ModuleProgress
	// Join with modules to filter by course
	err := r.db.Joins("JOIN modules ON modules.id = module_progresses.module_id").
		Where("module_progresses.user_id = ? AND modules.course_id = ?", userID, courseID).
		Find(&progresses).Error
	return progresses, err
}

func (r *courseRepository) AddPointsToProfile(userID uuid.UUID, points int) error {
	return r.db.Model(&model.Profile{}).Where("user_id = ?", userID).UpdateColumn("points", gorm.Expr("points + ?", points)).Error
}

func (r *courseRepository) LikeCourse(userID uuid.UUID, courseID uuid.UUID) error {
	like := model.CourseLike{
		UserID:   userID,
		CourseID: courseID,
	}
	// Use clause.OnConflict to ignore if it already exists
	return r.db.Save(&like).Error
}

func (r *courseRepository) UnlikeCourse(userID uuid.UUID, courseID uuid.UUID) error {
	return r.db.Where("user_id = ? AND course_id = ?", userID, courseID).Delete(&model.CourseLike{}).Error
}

func (r *courseRepository) HasUserLikedCourse(userID uuid.UUID, courseID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.CourseLike{}).Where("user_id = ? AND course_id = ?", userID, courseID).Count(&count).Error
	return count > 0, err
}

func (r *courseRepository) GetCourseLikesCount(courseID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.CourseLike{}).Where("course_id = ?", courseID).Count(&count).Error
	return count, err
}

func (r *courseRepository) GetTrendingCourses(limit int) ([]model.Course, error) {
	var courses []model.Course
	// Select courses joined with a count of course_likes
	err := r.db.
		Preload("Teacher.Profile").
		Preload("Enrollments").
		Joins("LEFT JOIN course_likes ON course_likes.course_id = courses.id").
		Select("courses.*, count(course_likes.user_id) as likes_count").
		Group("courses.id").
		Order("likes_count desc, created_at desc").
		Limit(limit).
		Find(&courses).Error

	return courses, err
}

func (r *courseRepository) CreateReview(review *model.CourseReview) error {
	return r.db.Create(review).Error
}

func (r *courseRepository) GetReviewsByCourseID(courseID uuid.UUID) ([]model.CourseReview, error) {
	var reviews []model.CourseReview
	err := r.db.Preload("User.Profile").Where("course_id = ?", courseID).Order("created_at desc").Find(&reviews).Error
	return reviews, err
}
