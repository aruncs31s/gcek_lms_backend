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
	GetAllCourses() ([]model.Course, error)
	UpdateCourse(course *model.Course) error
	DeleteCourse(id uuid.UUID) error

	CreateModule(module *model.Module) error
	GetModulesByCourseID(courseID uuid.UUID) ([]model.Module, error)
	GetMaxModuleOrderIndex(courseID uuid.UUID) int
	DeleteModule(id uuid.UUID) error
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
	err := r.db.Preload("Modules").First(&course, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &course, nil
}

func (r *courseRepository) GetAllCourses() ([]model.Course, error) {
	var courses []model.Course
	err := r.db.Find(&courses).Error
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

func (r *courseRepository) DeleteModule(id uuid.UUID) error {
	return r.db.Delete(&model.Module{}, "id = ?", id).Error
}
