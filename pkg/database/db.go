package database

import (
	"fmt"
	"log"

	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Kolkata",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v\n", err)
		return nil, err
	}

	err = db.AutoMigrate(
		&model.User{},
		&model.Profile{},
		&model.Course{},
		&model.Module{},
		&model.ModuleProgress{},
		&model.Enrollment{},
		&model.CourseReview{},
		&model.CourseLike{},
		&model.WatchLater{},
		&model.Conversation{},
		&model.ConversationParticipant{},
		&model.Message{},
		&model.Blog{},
		&model.Certificate{},
		&model.Assignment{},
		&model.AssignmentSubmission{},
		&model.Notification{},
		&model.Achievement{},
	)
	if err != nil {
		log.Printf("Failed to auto-migrate database: %v\n", err)
		return nil, err
	}

	log.Println("Connected to PostgreSQL database successfully.")
	return db, nil
}
