package main

import (
	"fmt"
	"log"

	"github.com/aruncs/esdc-lms/internal/handler"
	"github.com/aruncs/esdc-lms/internal/logger"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/aruncs/esdc-lms/internal/routes"
	"github.com/aruncs/esdc-lms/internal/service"
	"github.com/aruncs/esdc-lms/pkg/certgen"
	"github.com/aruncs/esdc-lms/pkg/config"
	"github.com/aruncs/esdc-lms/pkg/database"
	"github.com/aruncs/esdc-lms/pkg/ocr"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load Configuration
	cfg := config.LoadConfig()
	logger.Init(cfg.LogDir, cfg.LogLevel)
	// Connect to Database
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	achievementRepo := repository.NewAchievementRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	certRepo := repository.NewCertificateRepository(db)
	chatRepo := repository.NewChatRepository(db)
	assignmentRepo := repository.NewAssignmentRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	codingRepo := repository.NewCodingRepository(db)

	// Initialize Certificate Orchestrator
	orchestrator := certgen.NewOrchestrator("pkg/certgen/templates", "uploads/certificates")

	// Initialize Chat Hub
	chatHub := service.NewChatHub(chatRepo)
	go chatHub.Run()

	// Initialize Services
	jwtSecret := "supersecretkey_change_me_in_prod" // Replace with cfg.JWTSecret later
	userService := service.NewUserService(userRepo, achievementRepo, jwtSecret)
	courseService := service.NewCourseService(courseRepo)
	certService := service.NewCertificateService(certRepo, userRepo, courseRepo, orchestrator)
	chatService := service.NewChatService(chatRepo)
	ocrClient := ocr.NewClient("http://localhost:8000")
	notificationService := service.NewNotificationService(notificationRepo)
	assignmentService := service.NewAssignmentService(assignmentRepo, courseRepo, ocrClient, notificationService)
	codingService := service.NewCodingService(codingRepo, courseRepo)

	baseURL := cfg.MediaURL
	uploadDir := cfg.UploadDir
	// Initialize Handlers
	authHandler := handler.NewAuthHandler(userService)
	courseHandler := handler.NewCourseHandler(courseService)
	uploadHandler := handler.NewUploadHandler(uploadDir, baseURL)
	certHandler := handler.NewCertificateHandler(certService, baseURL)
	chatHandler := handler.NewChatHandler(chatHub, chatService)
	leaderboardHandler := handler.NewLeaderboardHandler(userRepo)
	assignmentHandler := handler.NewAssignmentHandler(assignmentService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	codingHandler := handler.NewCodingHandler(codingService)
	userHandler := handler.NewUserHandler(userService)

	fmt.Printf("Starting ESDC LMS Backend on port %s...\n", cfg.ServerPort)

	// Setup Gin Router
	r := gin.Default()

	// Setup CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	// Register all routes
	routes.SetupRoutes(
		r, jwtSecret,
		authHandler,
		courseHandler,
		uploadHandler,
		certHandler,
		chatHandler,
		leaderboardHandler,
		assignmentHandler,
		notificationHandler,
		codingHandler,
		userHandler,
	)

	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
