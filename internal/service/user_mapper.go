package service

import (
	"github.com/aruncs31s/gcek_lms_backend/internal/dto"
	"github.com/aruncs31s/gcek_lms_backend/pkg/model"
)

type UserMapper interface {
	MapToDTO(
		user *model.User,
	) *dto.UserResponse
	MapToDTOWithType(
		user *model.User,
	) *dto.UserResponseWithType
}
type userMapper struct{}

func NewUserMapper() UserMapper {
	return &userMapper{}
}

func (m *userMapper) MapToDTO(user *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID.String(),
		FirstName: user.Profile.FirstName,
		LastName:  user.Profile.LastName,
		Email:     user.Email,
		Role:      string(user.Role),
		AvatarURL: user.Profile.AvatarURL,
		Bio:       user.Profile.Bio,
	}
}

func (m *userMapper) MapToDTOWithType(user *model.User) *dto.UserResponseWithType {
	return &dto.UserResponseWithType{
		UserResponse: *m.MapToDTO(user),
		Type:         string(user.Role),
	}
}
