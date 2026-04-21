// Package mocks provides hand-written mock implementations of repository interfaces
// used for unit testing the certificate service without a real database.
package mocks

import (
	"context"
	"fmt"

	"github.com/aruncs31s/gcek_lms_backend/pkg/model"
	"github.com/google/uuid"
)

// ─── UserRepository mock ───────────────────────────────────────────────────

// MockUserRepository is a test double for repository.UserRepository.
type MockUserRepository struct {
	Users map[uuid.UUID]*model.User
	// GetUserByIDFunc lets tests override the default map lookup.
	GetUserByIDFunc func(id uuid.UUID) (*model.User, error)
}

func NewMockUserRepository(users ...*model.User) *MockUserRepository {
	m := &MockUserRepository{Users: make(map[uuid.UUID]*model.User)}
	for _, u := range users {
		m.Users[u.ID] = u
	}
	return m
}

func (m *MockUserRepository) GetUserByID(id uuid.UUID) (*model.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(id)
	}
	u, ok := m.Users[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return u, nil
}

func (m *MockUserRepository) CreateUser(user *model.User) error { return nil }
func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
	return nil, nil
}
func (m *MockUserRepository) GetLeaderboard(limit int) ([]model.User, error) {
	return nil, nil
}
func (m *MockUserRepository) List(limit, offset int, userType string) ([]model.User, int64, error) {
	return nil, 0, nil
}
func (m *MockUserRepository) GetProfileWithEnrolments(userID string, limit, offset int) (*model.User, []model.Enrollment, int64, error) {
	return nil, nil, 0, nil
}
func (m *MockUserRepository) UpdateProfile(profile *model.Profile) error { return nil }
func (m *MockUserRepository) Search(query string, role string, limit, offset int) ([]model.User, int64, error) {
	return nil, 0, nil
}
func (m *MockUserRepository) GetUsersByCourse(ctx context.Context, courseID uuid.UUID) ([]model.User, error) {
	return nil, nil
}
func (m *MockUserRepository) GetUserCountByCourse(ctx context.Context, courseID uuid.UUID) (int64, error) {
	return 0, nil
}

// ─── CourseRepository mock ─────────────────────────────────────────────────

// MockCourseRepository is a test double for repository.CourseRepository.
type MockCourseRepository struct {
	Courses map[uuid.UUID]*model.Course
	// GetCourseByIDFunc lets tests override the default map lookup.
	GetCourseByIDFunc func(id uuid.UUID) (*model.Course, error)

	// SaveCertificate fields
	SaveCertificateFunc func(cert *model.Certificate) error
	Calls               []*model.Certificate
}

func NewMockCourseRepository(courses ...*model.Course) *MockCourseRepository {
	m := &MockCourseRepository{Courses: make(map[uuid.UUID]*model.Course)}
	for _, c := range courses {
		m.Courses[c.ID] = c
	}
	return m
}

func (m *MockCourseRepository) GetCourseByID(id uuid.UUID) (*model.Course, error) {
	if m.GetCourseByIDFunc != nil {
		return m.GetCourseByIDFunc(id)
	}
	c, ok := m.Courses[id]
	if !ok {
		return nil, fmt.Errorf("course %s not found", id)
	}
	return c, nil
}

func (m *MockCourseRepository) SaveCertificate(cert *model.Certificate) error {
	m.Calls = append(m.Calls, cert)
	if m.SaveCertificateFunc != nil {
		return m.SaveCertificateFunc(cert)
	}
	if cert.ID == uuid.Nil {
		cert.ID = uuid.New()
	}
	return nil
}

func (m *MockCourseRepository) CreateCourse(course *model.Course) error { return nil }
func (m *MockCourseRepository) UpdateCourse(course *model.Course) error { return nil }
func (m *MockCourseRepository) DeleteCourse(id uuid.UUID) error         { return nil }
func (m *MockCourseRepository) CreateModule(module *model.Module) error { return nil }
func (m *MockCourseRepository) UpdateModule(module *model.Module) error { return nil }
func (m *MockCourseRepository) DeleteModule(id uuid.UUID) error         { return nil }
func (m *MockCourseRepository) UpdateModuleOrder(courseID uuid.UUID, orderedIDs []uuid.UUID) error {
	return nil
}
func (m *MockCourseRepository) CreateEnrollment(enrollment *model.Enrollment) error {
	return nil
}
func (m *MockCourseRepository) UpdateModuleProgress(progress *model.ModuleProgress) error {
	return nil
}
func (m *MockCourseRepository) LikeCourse(userID, courseID uuid.UUID) error   { return nil }
func (m *MockCourseRepository) UnlikeCourse(userID, courseID uuid.UUID) error { return nil }
func (m *MockCourseRepository) CreateReview(review *model.CourseReview) error {
	return nil
}
func (m *MockCourseRepository) GetAllCourses(query, courseType, format, status, teacherID string) ([]model.Course, error) {
	return nil, nil
}
func (m *MockCourseRepository) GetModulesByCourseID(courseID uuid.UUID) ([]model.Module, error) {
	return nil, nil
}
func (m *MockCourseRepository) GetMaxModuleOrderIndex(courseID uuid.UUID) int {
	return 0
}
func (m *MockCourseRepository) GetModuleByID(id uuid.UUID) (*model.Module, error) {
	return nil, nil
}
func (m *MockCourseRepository) GetEnrollment(userID, courseID uuid.UUID) (*model.Enrollment, error) {
	return nil, nil
}
func (m *MockCourseRepository) GetModuleProgresses(userID, courseID uuid.UUID) ([]model.ModuleProgress, error) {
	return nil, nil
}
func (m *MockCourseRepository) HasUserLikedCourse(userID, courseID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockCourseRepository) GetCourseLikesCount(courseID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockCourseRepository) GetTrendingCourses(limit int) ([]model.Course, error) {
	return nil, nil
}
func (m *MockCourseRepository) GetReviewsByCourseID(courseID uuid.UUID) ([]model.CourseReview, error) {
	return nil, nil
}
func (m *MockCourseRepository) AddPointsToProfile(userID uuid.UUID, points int) error {
	return nil
}
func (m *MockCourseRepository) SearchCourses(query, courseType, format, status string, limit, offset int) ([]model.Course, error) {
	return nil, nil
}
