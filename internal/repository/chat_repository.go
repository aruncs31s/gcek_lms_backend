package repository

import (
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRepository interface {
	CreateConversation(conv *model.Conversation) error
	GetConversationsByUserID(userID uuid.UUID) ([]model.Conversation, error)
	CreateMessage(msg *model.Message) error
	GetMessagesByConversationID(convID uuid.UUID) ([]model.Message, error)
	IsParticipant(convID, userID uuid.UUID) bool
	EnsureConversationExists(convID uuid.UUID) error
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db}
}

func (r *chatRepository) CreateConversation(conv *model.Conversation) error {
	return r.db.Create(conv).Error
}

func (r *chatRepository) GetConversationsByUserID(userID uuid.UUID) ([]model.Conversation, error) {
	var convs []model.Conversation
	err := r.db.Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id = ?", userID).
		Preload("Participants").
		Find(&convs).Error
	return convs, err
}

func (r *chatRepository) CreateMessage(msg *model.Message) error {
	return r.db.Create(msg).Error
}

func (r *chatRepository) GetMessagesByConversationID(convID uuid.UUID) ([]model.Message, error) {
	var msgs []model.Message
	err := r.db.Where("conversation_id = ?", convID).Order("created_at asc").Find(&msgs).Error
	return msgs, err
}

func (r *chatRepository) EnsureConversationExists(convID uuid.UUID) error {
	var count int64
	r.db.Model(&model.Conversation{}).Where("id = ?", convID).Count(&count)
	if count == 0 {
		return r.db.Create(&model.Conversation{
			ID:   convID,
			Type: model.ConversationTypeGroup,
		}).Error
	}
	return nil
}

func (r *chatRepository) IsParticipant(convID, userID uuid.UUID) bool {
	var count int64

	// First check if the user is enrolled in the course with this ID
	r.db.Model(&model.Enrollment{}).
		Where("course_id = ? AND user_id = ?", convID, userID).
		Count(&count)
	if count > 0 {
		return true
	}

	// Then check if the user is the teacher of the course with this ID
	r.db.Model(&model.Course{}).
		Where("id = ? AND teacher_id = ?", convID, userID).
		Count(&count)
	if count > 0 {
		return true
	}

	// Fallback to chat participant check
	r.db.Model(&model.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Count(&count)
	return count > 0
}
