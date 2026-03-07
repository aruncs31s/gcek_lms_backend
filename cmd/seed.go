package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/pkg/config"
	"github.com/aruncs/esdc-lms/pkg/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func RunSeed() {
	cfg := config.LoadConfig()
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Seed Teachers
	teachers := make([]model.User, 100)
	for i := 0; i < 100; i++ {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		teachers[i] = model.User{
			ID:           uuid.New(),
			Email:        fmt.Sprintf("teacher%d@gcek.edu", i+1),
			PasswordHash: string(hash),
			Role:         model.RoleTeacher,
		}
	}
	db.Create(&teachers)

	// Seed Teacher Profiles
	teacherProfiles := make([]model.Profile, 100)
	for i := 0; i < 100; i++ {
		teacherProfiles[i] = model.Profile{
			UserID:    teachers[i].ID,
			FirstName: fmt.Sprintf("Teacher%d", i+1),
			LastName:  fmt.Sprintf("LastName%d", i+1),
			Bio:       fmt.Sprintf("Experienced educator specializing in various subjects"),
			Points:    0,
		}
	}
	db.Create(&teacherProfiles)

	// Seed Students
	students := make([]model.User, 100)
	for i := 0; i < 100; i++ {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		students[i] = model.User{
			ID:           uuid.New(),
			Email:        fmt.Sprintf("student%d@gcek.edu", i+1),
			PasswordHash: string(hash),
			Role:         model.RoleStudent,
		}
	}
	db.Create(&students)

	// Seed Student Profiles
	studentProfiles := make([]model.Profile, 100)
	for i := 0; i < 100; i++ {
		studentProfiles[i] = model.Profile{
			UserID:    students[i].ID,
			FirstName: fmt.Sprintf("Student%d", i+1),
			LastName:  fmt.Sprintf("LastName%d", i+1),
			Bio:       fmt.Sprintf("Enthusiastic learner"),
			Points:    i * 10,
		}
	}
	db.Create(&studentProfiles)

	// Seed Courses
	courses := make([]model.Course, 100)
	startDate := time.Now()
	for i := 0; i < 100; i++ {
		courses[i] = model.Course{
			ID:                     uuid.New(),
			TeacherID:              teachers[i%100].ID,
			Title:                  fmt.Sprintf("Course %d: Advanced Topics", i+1),
			Description:            fmt.Sprintf("Comprehensive course covering advanced concepts in subject area %d", i+1),
			Price:                  float64((i%10 + 1) * 100),
			Type:                   []string{"free", "paid"}[i%2],
			Format:                 "course",
			Status:                 []string{"active", "coming soon"}[i%2],
			Duration:               fmt.Sprintf("%d weeks", (i%12)+1),
			IsCertificateAvailable: i%2 == 0,
			StartDate:              &startDate,
		}
	}
	db.Create(&courses)

	// Seed Projects
	projects := make([]model.Course, 100)
	for i := 0; i < 100; i++ {
		projects[i] = model.Course{
			ID:                     uuid.New(),
			TeacherID:              teachers[i%100].ID,
			Title:                  fmt.Sprintf("Project %d: Real-World Application", i+1),
			Description:            fmt.Sprintf("Hands-on project focusing on practical implementation of concepts %d", i+1),
			Price:                  float64((i%5 + 1) * 50),
			Type:                   []string{"free", "paid"}[i%3],
			Format:                 "project",
			Status:                 "active",
			Duration:               fmt.Sprintf("%d weeks", (i%8)+2),
			IsCertificateAvailable: true,
			StartDate:              &startDate,
		}
	}
	db.Create(&projects)
}
