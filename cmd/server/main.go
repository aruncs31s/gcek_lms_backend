package main

import (
	"fmt"
	"log"

	"github.com/aruncs/esdc-lms/internal/handler"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
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

	baseURL := "http://localhost:" + cfg.ServerPort // Should ideally be from config

	// Initialize Handlers
	authHandler := handler.NewAuthHandler(userService)
	courseHandler := handler.NewCourseHandler(courseService)
	uploadHandler := handler.NewUploadHandler(baseURL)
	certHandler := handler.NewCertificateHandler(certService, baseURL)
	chatHandler := handler.NewChatHandler(chatHub, chatService)
	leaderboardHandler := handler.NewLeaderboardHandler(userRepo)
	assignmentHandler := handler.NewAssignmentHandler(assignmentService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	fmt.Printf("Starting ESDC LMS Backend on port %s...\n", cfg.ServerPort)

	// Setup Gin Router
	r := gin.Default()

	// Setup CORS
	configCors := cors.DefaultConfig()
	configCors.AllowAllOrigins = true
	configCors.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(configCors))

	// Global Middlewares
	authMw := middleware.AuthMiddleware(jwtSecret)
	optAuthMw := middleware.OptionalAuthMiddleware(jwtSecret)
	// Now allowing Student, Teacher, Admin to access course creation routes
	teacherAdminMw := middleware.RoleMiddleware(string(model.RoleTeacher), string(model.RoleAdmin), string(model.RoleStudent))

	// Static Files
	r.Static("/uploads", "./uploads")

	// Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Setup API Routes
	api := r.Group("/api")
	{
		// Auth Routes
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		// Leaderboard Routes
		api.GET("/leaderboard", leaderboardHandler.GetLeaderboard)

		// Course Routes
		courses := api.Group("/courses")
		courses.Use(optAuthMw)
		{
			courses.GET("", courseHandler.GetAllCourses)
			courses.GET("/:id", courseHandler.GetCourseByID)
		}

		protectedCourses := api.Group("/courses") // Recreate without optAuthMw to prevent duplicate middleware runs
		protectedCourses.Use(authMw)
		{
			protectedCourses.POST("/:id/enroll", courseHandler.EnrollCourse)
			protectedCourses.GET("/:id/enrollment", courseHandler.GetEnrollmentStatus)
			protectedCourses.POST("/:id/modules/:moduleId/complete", courseHandler.CompleteModule)

			// Teacher only
			teacherCourses := protectedCourses.Group("")
			teacherCourses.Use(teacherAdminMw)
			{
				teacherCourses.POST("", courseHandler.CreateCourse)
				teacherCourses.PUT("/:id", courseHandler.UpdateCourse)
				teacherCourses.PATCH("/:id", courseHandler.UpdateCourse)
				teacherCourses.DELETE("/:id", courseHandler.DeleteCourse)
				teacherCourses.POST("/:id/modules", courseHandler.CreateModule)
				teacherCourses.PUT("/:id/modules/:moduleId", courseHandler.UpdateModule)
				teacherCourses.PATCH("/:id/modules/:moduleId", courseHandler.UpdateModule)
				teacherCourses.PUT("/:id/modules/reorder", courseHandler.ReorderModules)
				teacherCourses.DELETE("/:id/modules/:moduleId", courseHandler.DeleteModule)

				// Assignment specific Teacher routes
				teacherCourses.POST("/:id/assignments", assignmentHandler.CreateAssignment)
				teacherCourses.PUT("/:id/assignments/:assignmentId", assignmentHandler.UpdateAssignment)
				teacherCourses.DELETE("/:id/assignments/:assignmentId", assignmentHandler.DeleteAssignment)
				teacherCourses.GET("/:id/assignments/:assignmentId/submissions", assignmentHandler.GetSubmissions)
				teacherCourses.PUT("/:id/assignments/:assignmentId/submissions/:submissionId/grade", assignmentHandler.GradeSubmission)
			}

			// Assignment general routes (for Enrolled and Teachers)
			protectedCourses.GET("/:id/assignments", assignmentHandler.GetAssignments)
			protectedCourses.GET("/:id/assignments/:assignmentId", assignmentHandler.GetAssignmentByID)
			protectedCourses.GET("/:id/assignments/:assignmentId/submissions/me", assignmentHandler.GetStudentSubmission)
			protectedCourses.POST("/:id/assignments/:assignmentId/submit", assignmentHandler.SubmitAssignment)
		}

		// Upload Routes
		upload := api.Group("/upload")
		upload.Use(authMw, teacherAdminMw)
		{
			upload.POST("/video", uploadHandler.UploadVideo)
			upload.POST("/image", uploadHandler.UploadImage)
		}

		// Certificate Routes
		certs := api.Group("/certificates")
		certs.Use(authMw)
		{
			certs.POST("/generate", certHandler.GenerateCertificate)
		}

		// Chat API Routes
		chat := api.Group("/chat")
		chat.Use(authMw)
		{
			chat.GET("/conversations", chatHandler.GetConversations)
			chat.POST("/conversations", chatHandler.CreateConversation)
			chat.GET("/conversations/:id/messages", chatHandler.GetMessages)
		}

		// Notification Routes
		notifications := api.Group("/notifications")
		notifications.Use(authMw)
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
		}
	}

	// WebSocket Routes
	r.GET("/ws/chat", chatHandler.ServeWS)
	userHandler := handler.NewUserHandler(userService)

	user := api.Group("/users")
	user.Use(authMw)
	{
		user.GET("", userHandler.List)
		user.GET("/:id/enrolments", userHandler.Enrolments)
	}
	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
