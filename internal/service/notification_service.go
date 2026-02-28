package service

import (
	"time"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/google/uuid"
)

type NotificationService interface {
	CreateNotification(userID uuid.UUID, title, message, notifType string) error
	GetUserNotifications(userID uuid.UUID) ([]dto.NotificationResponse, error)
	MarkAsRead(notificationID, userID uuid.UUID) error
	GetUnreadCount(userID uuid.UUID) (int64, error)
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
}

func NewNotificationService(notificationRepo repository.NotificationRepository) NotificationService {
	return &notificationService{notificationRepo: notificationRepo}
}

func (s *notificationService) CreateNotification(userID uuid.UUID, title, message, notifType string) error {
	notification := &model.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
	}
	return s.notificationRepo.CreateNotification(notification)
}

func (s *notificationService) GetUserNotifications(userID uuid.UUID) ([]dto.NotificationResponse, error) {
	notifications, err := s.notificationRepo.GetNotificationsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responses []dto.NotificationResponse
	for _, n := range notifications {
		responses = append(responses, dto.NotificationResponse{
			ID:        n.ID,
			Title:     n.Title,
			Message:   n.Message,
			IsRead:    n.IsRead,
			Type:      n.Type,
			CreatedAt: n.CreatedAt.Format(time.RFC3339),
		})
	}
	if responses == nil {
		responses = []dto.NotificationResponse{} // Ensure empty array, not null
	}
	return responses, nil
}

func (s *notificationService) MarkAsRead(notificationID, userID uuid.UUID) error {
	return s.notificationRepo.MarkAsRead(notificationID, userID)
}

func (s *notificationService) GetUnreadCount(userID uuid.UUID) (int64, error) {
	return s.notificationRepo.GetUnreadCount(userID)
}
