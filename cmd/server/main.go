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
	courseRepo := repository.NewCourseRepository(db)
	certRepo := repository.NewCertificateRepository(db)
	chatRepo := repository.NewChatRepository(db)

	// Initialize Certificate Orchestrator
	orchestrator := certgen.NewOrchestrator("pkg/certgen/templates", "uploads/certificates")

	// Initialize Chat Hub
	chatHub := service.NewChatHub(chatRepo)
	go chatHub.Run()

	// Initialize Services
	jwtSecret := "supersecretkey_change_me_in_prod" // Replace with cfg.JWTSecret later
	userService := service.NewUserService(userRepo, jwtSecret)
	courseService := service.NewCourseService(courseRepo)
	certService := service.NewCertificateService(certRepo, userRepo, courseRepo, orchestrator)
	chatService := service.NewChatService(chatRepo)

	baseURL := "http://localhost:" + cfg.ServerPort // Should ideally be from config

	// Initialize Handlers
	authHandler := handler.NewAuthHandler(userService)
	courseHandler := handler.NewCourseHandler(courseService)
	uploadHandler := handler.NewUploadHandler(baseURL)
	certHandler := handler.NewCertificateHandler(certService, baseURL)
	chatHandler := handler.NewChatHandler(chatHub, chatService)

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
	teacherAdminMw := middleware.RoleMiddleware(string(model.RoleTeacher), string(model.RoleAdmin))

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

		// Course Routes
		courses := api.Group("/courses")
		{
			// Public (or just authenticated depending on logic)
			courses.GET("", courseHandler.GetAllCourses)
			courses.GET("/:id", courseHandler.GetCourseByID)

			// Protected
			courses.Use(authMw)
			courses.Use(teacherAdminMw)
			courses.POST("", courseHandler.CreateCourse)
			courses.PUT("/:id", courseHandler.UpdateCourse)
			courses.PATCH("/:id", courseHandler.UpdateCourse)
			courses.DELETE("/:id", courseHandler.DeleteCourse)
			courses.POST("/:id/modules", courseHandler.CreateModule)
			courses.DELETE("/:id/modules/:moduleId", courseHandler.DeleteModule)
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
	}

	// WebSocket Routes
	r.GET("/ws/chat", chatHandler.ServeWS)

	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
