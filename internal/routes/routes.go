package routes

import (
	"github.com/aruncs/esdc-lms/internal/handler"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	jwtSecret string,
	authHandler *handler.AuthHandler,
	courseHandler *handler.CourseHandler,
	uploadHandler handler.UploadHandler,
	certHandler *handler.CertificateHandler,
	chatHandler *handler.ChatHandler,
	leaderboardHandler *handler.LeaderboardHandler,
	assignmentHandler *handler.AssignmentHandler,
	notificationHandler *handler.NotificationHandler,
	codingHandler *handler.CodingHandler,
	userHandler *handler.UserHandler,
) {
	// Middlewares
	authMw := middleware.AuthMiddleware(jwtSecret)
	optAuthMw := middleware.OptionalAuthMiddleware(jwtSecret)
	teacherAdminMw := middleware.RoleMiddleware(string(model.RoleTeacher), string(model.RoleAdmin), string(model.RoleStudent))

	// Static Files
	r.Static("/uploads", "./uploads")

	// Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// WebSocket Routes
	r.GET("/ws/chat", chatHandler.ServeWS)

	api := r.Group("/api")
	{
		// Auth Routes
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		// Leaderboard Routes
		api.GET("/leaderboard", leaderboardHandler.GetLeaderboard)

		// Course Routes (public / optional auth)
		courses := api.Group("/courses")
		courses.Use(optAuthMw)
		{
			courses.GET("", courseHandler.GetAllCourses)
			courses.GET("/trending", courseHandler.GetTrendingCourses)
			courses.GET("/search", courseHandler.SearchCourses)
			courses.GET("/:id", courseHandler.GetCourseByID)
			courses.GET("/:id/reviews", courseHandler.GetReviews)
		}

		// Course Routes (protected)
		protectedCourses := api.Group("/courses")
		protectedCourses.Use(authMw)
		{
			protectedCourses.POST("/:id/like", courseHandler.LikeCourse)
			protectedCourses.DELETE("/:id/like", courseHandler.UnlikeCourse)
			protectedCourses.POST("/:id/enroll", courseHandler.EnrollCourse)
			protectedCourses.GET("/:id/enrollment", courseHandler.GetEnrollmentStatus)
			protectedCourses.POST("/:id/modules/:moduleId/complete", courseHandler.CompleteModule)
			protectedCourses.POST("/:id/reviews", courseHandler.AddReview)

			// Teacher / Admin only
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

				// Assignment – teacher routes
				teacherCourses.POST("/:id/assignments", assignmentHandler.CreateAssignment)
				teacherCourses.PUT("/:id/assignments/:assignmentId", assignmentHandler.UpdateAssignment)
				teacherCourses.DELETE("/:id/assignments/:assignmentId", assignmentHandler.DeleteAssignment)
				teacherCourses.GET("/:id/assignments/:assignmentId/submissions", assignmentHandler.GetSubmissions)
				teacherCourses.PUT("/:id/assignments/:assignmentId/submissions/:submissionId/grade", assignmentHandler.GradeSubmission)

				// Coding Assignment – teacher routes
				teacherCourses.POST("/:id/coding-assignments", codingHandler.CreateCodingAssignment)
				teacherCourses.PUT("/:id/coding-assignments/:codingAssignmentId", codingHandler.UpdateCodingAssignment)
				teacherCourses.DELETE("/:id/coding-assignments/:codingAssignmentId", codingHandler.DeleteCodingAssignment)
				teacherCourses.GET("/:id/coding-assignments/:codingAssignmentId/submissions", codingHandler.GetSubmissions)
				teacherCourses.PUT("/:id/coding-assignments/:codingAssignmentId/submissions/:submissionId/grade", codingHandler.GradeSubmission)
			}

			// Assignment – general routes (enrolled students + teachers)
			protectedCourses.GET("/:id/assignments", assignmentHandler.GetAssignments)
			protectedCourses.GET("/:id/assignments/:assignmentId", assignmentHandler.GetAssignmentByID)
			protectedCourses.GET("/:id/assignments/:assignmentId/submissions/me", assignmentHandler.GetStudentSubmission)
			protectedCourses.POST("/:id/assignments/:assignmentId/submit", assignmentHandler.SubmitAssignment)

			// Coding Assignment – general routes (students + teachers)
			protectedCourses.GET("/:id/coding-assignments", codingHandler.GetCodingAssignments)
			protectedCourses.GET("/:id/coding-assignments/:codingAssignmentId", codingHandler.GetCodingAssignmentByID)
			protectedCourses.POST("/:id/coding-assignments/:codingAssignmentId/submit", codingHandler.SubmitCode)
			protectedCourses.GET("/:id/coding-assignments/:codingAssignmentId/submissions/me", codingHandler.GetMySubmission)
		}

		// Code Execution (sandbox – auth required)
		codeRun := api.Group("/code")
		codeRun.Use(authMw)
		{
			codeRun.POST("/run", codingHandler.RunCode)
		}

		// Upload Routes
		upload := api.Group("/upload")
		upload.Use(authMw)
		{
			upload.POST("/video", teacherAdminMw, uploadHandler.UploadVideo)
			upload.POST("/image", uploadHandler.UploadImage)
			upload.POST("/attachment", uploadHandler.UploadAttachment)
		}

		// Certificate Routes
		certs := api.Group("/certificates")
		certs.Use(authMw)
		{
			certs.POST("/generate", certHandler.GenerateCertificate)
		}

		// Chat Routes
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

		// User Routes
		user := api.Group("/users")
		user.Use(authMw)
		{
			user.GET("", userHandler.List)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.GET("/:id/enrolments", userHandler.Enrolments)
		}
	}
}
