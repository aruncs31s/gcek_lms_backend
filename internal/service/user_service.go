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
	List(limit, offset int, userType string) (users []dto.UserResponseWithType, count int64, err error)
	Enrolments(limit, offset int, userId string) (*dto.UserProfileEnrolmentsResponse, error)
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
		},
		Points:       user.Profile.Points,
		Achievements: achievementsRes,
		Enrolments:   enrolmentsRes,
		TotalCount:   count,
	}, nil
}
