package repository

import (
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AchievementRepository interface {
	CreateAchievement(achievement *model.Achievement) error
	FindByUserID(userID uuid.UUID) ([]model.Achievement, error)
}

type achievementRepository struct {
	db *gorm.DB
}

func NewAchievementRepository(db *gorm.DB) AchievementRepository {
	return &achievementRepository{db: db}
}

func (r *achievementRepository) CreateAchievement(achievement *model.Achievement) error {
	return r.db.Create(achievement).Error
}

func (r *achievementRepository) FindByUserID(userID uuid.UUID) ([]model.Achievement, error) {
	var achievements []model.Achievement
	err := r.db.Where("user_id = ?", userID).Find(&achievements).Error
	return achievements, err
}
