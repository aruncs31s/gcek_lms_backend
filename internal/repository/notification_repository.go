package repository

import (
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	CreateNotification(notification *model.Notification) error
	GetNotificationsByUserID(userID uuid.UUID) ([]model.Notification, error)
	MarkAsRead(notificationID, userID uuid.UUID) error
	GetUnreadCount(userID uuid.UUID) (int64, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) CreateNotification(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetNotificationsByUserID(userID uuid.UUID) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) MarkAsRead(notificationID, userID uuid.UUID) error {
	return r.db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}
