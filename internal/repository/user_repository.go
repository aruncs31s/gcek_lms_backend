package repository

import (
	"errors"

	"github.com/aruncs/esdc-lms/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	GetLeaderboard(limit int) ([]model.User, error)
	List(limit, offset int, userType string) ([]model.User, int64, error)
	GetProfileWithEnrolments(userID string, limit, offset int) (*model.User, []model.Enrollment, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Profile").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil to indicate not found clearly
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetLeaderboard(limit int) ([]model.User, error) {
	var users []model.User
	err := r.db.
		Preload("Profile").
		Preload("Enrollments").
		Joins("JOIN profiles ON profiles.user_id = users.id").
		Order("profiles.points desc").
		Limit(limit).
		Find(&users).Error
	return users, err
}
func (r *userRepository) List(limit, offset int, userType string) ([]model.User, int64, error) {
	var users []model.User
	var count int64
	if userType == "" || userType == "all" {
		err := r.db.Preload("Profile").Offset(offset).Limit(limit).Find(&users).Count(&count).Error
		return users, count, err
	}
	err := r.db.Preload("Profile").Where("role = ?", userType).Offset(offset).Limit(limit).Find(&users).Count(&count).Error
	return users, count, err
}

func (r *userRepository) GetProfileWithEnrolments(userID string, limit, offset int) (*model.User, []model.Enrollment, int64, error) {
	var user model.User
	err := r.db.Preload("Profile").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, nil, 0, err
	}

	var enrollments []model.Enrollment
	var count int64
	err = r.db.Model(&model.Enrollment{}).Preload("Course").Where("user_id = ?", userID).Count(&count).Offset(offset).Limit(limit).Find(&enrollments).Error

	return &user, enrollments, count, err
}
