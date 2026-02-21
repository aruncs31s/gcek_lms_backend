package service

import (
	"errors"
	"time"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	RegisterUser(req *dto.RegisterRequest) (*dto.AuthResponse, error)
	LoginUser(req *dto.LoginRequest) (*dto.AuthResponse, error)
}

type userService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

func NewUserService(repo repository.UserRepository, secret string) UserService {
	return &userService{
		userRepo:  repo,
		jwtSecret: []byte(secret),
	}
}

func (s *userService) RegisterUser(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user exists
	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already in use")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := uuid.New()
	user := &model.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         model.Role(req.Role),
		Profile: model.Profile{
			UserID:    userID,
			FirstName: req.FirstName,
			LastName:  req.LastName,
		},
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID.String(),
			FirstName: user.Profile.FirstName,
			LastName:  user.Profile.LastName,
			Email:     user.Email,
			Role:      string(user.Role),
			AvatarURL: user.Profile.AvatarURL,
		},
	}, nil
}

func (s *userService) LoginUser(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.generateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID.String(),
			FirstName: user.Profile.FirstName,
			LastName:  user.Profile.LastName,
			Email:     user.Email,
			Role:      string(user.Role),
			AvatarURL: user.Profile.AvatarURL,
		},
	}, nil
}

func (s *userService) generateToken(userID uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
