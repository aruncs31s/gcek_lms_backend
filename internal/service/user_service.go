package service

import (
	"errors"
	"time"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/logger"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAlreadyExists      = errors.New("email already in use")
)

type UserService interface {
	RegisterUser(
		req *dto.RegisterRequest,
	) (*dto.AuthResponse, error)
	LoginUser(
		req *dto.LoginRequest,
	) (*dto.AuthResponse, error)
	List(
		limit,
		offset int,
		userType string,
	) (users []dto.UserResponseWithType, count int64, err error)
	Enrolments(
		limit,
		offset int,
		userID string) (*dto.UserProfileEnrolmentsResponse, error)
	UpdateProfile(
		userID string,
		req *dto.UpdateProfileRequest,
	) (*dto.UserResponse, error)
	Search(
		query string,
		role string,
		limit,
		offset int,
	) (users []dto.UserResponse, count int64, err error)
}

type userService struct {
	userRepo        repository.UserRepository
	achievementRepo repository.AchievementRepository
	jwtSecret       []byte
}

func NewUserService(repo repository.UserRepository, achievementRepo repository.AchievementRepository, secret string) UserService {
	return &userService{
		userRepo:        repo,
		achievementRepo: achievementRepo,
		jwtSecret:       []byte(secret),
	}
}

func (s *userService) RegisterUser(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user exists
	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, ErrAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.GetLogger().Error(
			"Failed to hash password: ",
			zap.String("err", err.Error()),
		)
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
		logger.GetLogger().Error(
			"Failed to create user: ",
			zap.String("err", err.Error()),
		)
		return nil, err
	}

	token, err := s.generateToken(user.ID, string(user.Role))
	if err != nil {
		logger.GetLogger().Error(
			"Failed to generate token: ",
			zap.String("err", err.Error()),
		)
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
			Bio:       user.Profile.Bio,
		},
	}, nil
}

func (s *userService) LoginUser(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
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
			Bio:       user.Profile.Bio,
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

func (s *userService) List(limit, offset int, userType string) ([]dto.UserResponseWithType, int64, error) {
	users, count, err := s.userRepo.List(limit, offset, userType)
	if err != nil {
		return nil, 0, err
	}
	var res []dto.UserResponseWithType
	for _, u := range users {
		res = append(res, dto.UserResponseWithType{
			UserResponse: dto.UserResponse{
				ID:        u.ID.String(),
				FirstName: u.Profile.FirstName,
				LastName:  u.Profile.LastName,
				Email:     u.Email,
				Role:      string(u.Role),
				AvatarURL: u.Profile.AvatarURL,
				Bio:       u.Profile.Bio,
			},
			Type: string(u.Role),
		})
	}
	return res, count, nil
}

func (s *userService) Enrolments(limit, offset int, userId string) (*dto.UserProfileEnrolmentsResponse, error) {
	// 1. Fetch user (with Profile) and enrollments (with Course)
	user, enrollments, count, err := s.userRepo.GetProfileWithEnrolments(userId, limit, offset)
	if err != nil {
		return nil, err
	}

	// 2. Fetch user's achievements
	achievements, err := s.achievementRepo.FindByUserID(user.ID)
	if err != nil {
		// Just log or ignore achievement errors if we want, or return
		return nil, err
	}

	// 3. Map to DTO
	var enrolmentsRes []dto.EnrolmentResponse
	for _, e := range enrollments {
		enrolmentsRes = append(enrolmentsRes, dto.EnrolmentResponse{
			CourseID:           e.CourseID.String(),
			CourseTitle:        e.Course.Title,
			CourseThumbnailURL: e.Course.ThumbnailURL,
			Status:             string(e.Status),
			ProgressPercentage: e.ProgressPercentage,
			EnrolledAt:         e.EnrolledAt,
		})
	}

	var achievementsRes []dto.AchievementResponse
	for _, a := range achievements {
		achievementsRes = append(achievementsRes, dto.AchievementResponse{
			ID:          a.ID.String(),
			Title:       a.Title,
			Description: a.Description,
			IconURL:     a.IconURL,
			Points:      a.Points,
			EarnedAt:    a.CreatedAt,
		})
	}

	return &dto.UserProfileEnrolmentsResponse{
		User: dto.UserResponse{
			ID:        user.ID.String(),
			FirstName: user.Profile.FirstName,
			LastName:  user.Profile.LastName,
			Email:     user.Email,
			Role:      string(user.Role),
			AvatarURL: user.Profile.AvatarURL,
			Bio:       user.Profile.Bio,
		},
		Points:       user.Profile.Points,
		Achievements: achievementsRes,
		Enrolments:   enrolmentsRes,
		TotalCount:   count,
	}, nil
}

func (s *userService) UpdateProfile(userId string, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, _, _, err := s.userRepo.GetProfileWithEnrolments(userId, 1, 0)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Profile.FirstName = req.FirstName
	user.Profile.LastName = req.LastName
	user.Profile.Bio = req.Bio
	if req.AvatarURL != "" {
		user.Profile.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.UpdateProfile(&user.Profile); err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.ID.String(),
		FirstName: user.Profile.FirstName,
		LastName:  user.Profile.LastName,
		Email:     user.Email,
		Role:      string(user.Role),
		AvatarURL: user.Profile.AvatarURL,
		Bio:       user.Profile.Bio,
	}, nil
}

func (s *userService) Search(query string, role string, limit, offset int) ([]dto.UserResponse, int64, error) {
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	if valid := model.Role(role).IsValid(); role != "" && !valid {
		return nil, 0, errors.New("invalid role filter")
	}
	users, count, err := s.userRepo.Search(query, role, limit, offset)
	if err != nil {
		logger.GetLogger().Error(
			"Failed to search users",
			zap.String("query", query),
			zap.String("err", err.Error()),
		)
		return nil, 0, err
	}

	var result []dto.UserResponse
	for _, user := range users {
		result = append(result, dto.UserResponse{
			ID:        user.ID.String(),
			FirstName: user.Profile.FirstName,
			LastName:  user.Profile.LastName,
			Email:     user.Email,
			Role:      string(user.Role),
			AvatarURL: user.Profile.AvatarURL,
			Bio:       user.Profile.Bio,
		})
	}

	return result, count, nil
}
