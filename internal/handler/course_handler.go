package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/aruncs/esdc-lms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CourseHandler struct {
	courseService service.CourseService
}

func NewCourseHandler(courseService service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

// CreateCourse godoc
// @Summary      Create a course
// @Description  Creates a new course. Requires Teacher or Admin role.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateCourseRequest  true  "Course creation payload"
// @Success      201   {object}  dto.CourseResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses [post]
func (h *CourseHandler) CreateCourse(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	teacherID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req dto.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.courseService.CreateCourse(teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetAllCourses godoc
// @Summary      List all courses
// @Description  Returns a list of courses with optional filters.
// @Tags         courses
// @Produce      json
// @Param        query      query  string  false  "Search query"
// @Param        type       query  string  false  "Course type (free, paid)"
// @Param        format     query  string  false  "Course format (course, project)"
// @Param        status     query  string  false  "Course status (coming soon, active, ended)"
// @Param        teacher_id query  string  false  "Filter by teacher ID (UUID)"
// @Success      200  {array}   dto.CourseResponse
// @Failure      500  {object}  map[string]string
// @Router       /api/courses [get]
func (h *CourseHandler) GetAllCourses(c *gin.Context) {
	var userID uuid.UUID
	if claimsRaw, exists := c.Get(middleware.UserContextKey); exists {
		if claims, ok := claimsRaw.(middleware.UserClaims); ok {
			userID, _ = uuid.Parse(claims.UserID)
		}
	}

	query := c.Query("query")
	courseType := c.Query("type")
	format := c.Query("format")
	status := c.Query("status")
	teacherID := c.Query("teacher_id")

	courses, err := h.courseService.GetAllCourses(userID, query, courseType, format, status, teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, courses)
}

// GetCourseByID godoc
// @Summary      Get course by ID
// @Description  Returns detailed information about a specific course, including modules.
// @Tags         courses
// @Produce      json
// @Param        id  path      string  true  "Course ID (UUID)"
// @Success      200  {object}  dto.CourseResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/courses/{id} [get]
func (h *CourseHandler) GetCourseByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var userID uuid.UUID
	if claimsRaw, exists := c.Get(middleware.UserContextKey); exists {
		if claims, ok := claimsRaw.(middleware.UserClaims); ok {
			userID, _ = uuid.Parse(claims.UserID)
		}
	}

	course, err := h.courseService.GetCourseByID(id, userID)
	if err != nil {
		if err.Error() == "course not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, course)
}

// UpdateCourse godoc
// @Summary      Update a course
// @Description  Updates an existing course. Requires Teacher or Admin role.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        id    path      string                   true  "Course ID (UUID)"
// @Param        body  body      dto.UpdateCourseRequest  true  "Course update payload"
// @Success      200   {object}  dto.CourseResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id} [put]
func (h *CourseHandler) UpdateCourse(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	teacherID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	idStr := c.Param("id")
	courseID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var req dto.UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.courseService.UpdateCourse(courseID, teacherID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "course not found" {
			status = http.StatusNotFound
		} else if strings.HasPrefix(err.Error(), "forbidden") {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// DeleteCourse godoc
// @Summary      Delete a course
// @Description  Deletes a course. Requires Teacher or Admin role.
// @Tags         courses
// @Produce      json
// @Param        id  path  string  true  "Course ID (UUID)"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id} [delete]
func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	teacherID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	idStr := c.Param("id")
	courseID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	err = h.courseService.DeleteCourse(courseID, teacherID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "course not found" {
			status = http.StatusNotFound
		} else if strings.HasPrefix(err.Error(), "forbidden") {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateModule godoc
// @Summary      Create a module
// @Description  Adds a new module to a course. Requires Teacher or Admin role.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        id    path      string                   true  "Course ID (UUID)"
// @Param        body  body      dto.CreateModuleRequest  true  "Module creation payload"
// @Success      201   {object}  dto.ModuleResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/modules [post]
func (h *CourseHandler) CreateModule(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	teacherID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var req dto.CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.courseService.CreateModule(courseID, teacherID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "course not found" {
			status = http.StatusNotFound
		} else if strings.HasPrefix(err.Error(), "forbidden") {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// DeleteModule godoc
// @Summary      Delete a module
// @Description  Removes a module from a course. Requires Teacher or Admin role.
// @Tags         courses
// @Produce      json
// @Param        id        path  string  true  "Course ID (UUID)"
// @Param        moduleId  path  string  true  "Module ID (UUID)"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/modules/{moduleId} [delete]
func (h *CourseHandler) DeleteModule(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	teacherID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	moduleIDStr := c.Param("moduleId")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	err = h.courseService.DeleteModule(moduleID, teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateModule godoc
// @Summary      Update a module
// @Description  Updates an existing module. Requires Teacher or Admin role.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        id        path      string                   true  "Course ID (UUID)"
// @Param        moduleId  path      string                   true  "Module ID (UUID)"
// @Param        body      body      dto.UpdateModuleRequest  true  "Module update payload"
// @Success      200   {object}  dto.ModuleResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/modules/{moduleId} [put]
func (h *CourseHandler) UpdateModule(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	teacherID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	moduleIDStr := c.Param("moduleId")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	var req dto.UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.courseService.UpdateModule(courseID, moduleID, teacherID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "course not found" || err.Error() == "module not found in the specified course" {
			status = http.StatusNotFound
		} else if strings.HasPrefix(err.Error(), "forbidden") {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// ReorderModules godoc
// @Summary      Reorder modules
// @Description  Sets the display order of modules within a course. Requires Teacher or Admin role.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "Course ID (UUID)"
// @Param        body  body      dto.ReorderModulesRequest true  "Ordered list of module IDs"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/modules/reorder [put]
func (h *CourseHandler) ReorderModules(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	teacherID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var req dto.ReorderModulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.courseService.ReorderModules(courseID, teacherID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "course not found" {
			status = http.StatusNotFound
		} else if strings.HasPrefix(err.Error(), "forbidden") {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Modules reordered successfully"})
}

// EnrollCourse godoc
// @Summary      Enroll in a course
// @Description  Enrolls the authenticated user in the specified course.
// @Tags         courses
// @Produce      json
// @Param        id  path  string  true  "Course ID (UUID)"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/enroll [post]
func (h *CourseHandler) EnrollCourse(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	err = h.courseService.EnrollCourse(courseID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully enrolled"})
}

// GetEnrollmentStatus godoc
// @Summary      Get enrollment status
// @Description  Returns the authenticated user's enrollment status for a course.
// @Tags         courses
// @Produce      json
// @Param        id  path  string  true  "Course ID (UUID)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/enrollment [get]
func (h *CourseHandler) GetEnrollmentStatus(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	enrollment, err := h.courseService.GetEnrollmentStatus(courseID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if enrollment == nil {
		c.JSON(http.StatusOK, gin.H{"enrolled": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enrolled":            true,
		"status":              enrollment.Status,
		"progress_percentage": enrollment.ProgressPercentage,
	})
}

func (h *CourseHandler) GetEntrolledUsers(
	c *gin.Context,
) {
	userID, courseID := h.getEssentials(c)
	if userID == uuid.Nil || courseID == uuid.Nil {
		return
	}

	enrollments, err := h.courseService.GetEnrolledUsers(c.Request.Context(), courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]any{"enrollments": enrollments})
}

func (*CourseHandler) getEssentials(c *gin.Context) (uuid.UUID, uuid.UUID) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return uuid.UUID{}, uuid.UUID{}
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return uuid.UUID{}, uuid.UUID{}
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return uuid.UUID{}, uuid.UUID{}
	}
	return userID, courseID
}

func (h *CourseHandler) GetEntrolledUsersCount(
	c *gin.Context,
) {
	courseId, userId := h.getEssentials(c)
	if courseId == uuid.Nil || userId == uuid.Nil {
		return
	}

	count, err := h.courseService.GetEnrolledUsersCount(c.Request.Context(), courseId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// CompleteModule godoc
// @Summary      Mark module as complete
// @Description  Marks a specific module as completed for the authenticated user.
// @Tags         courses
// @Produce      json
// @Param        id        path  string  true  "Course ID (UUID)"
// @Param        moduleId  path  string  true  "Module ID (UUID)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/modules/{moduleId}/complete [post]
func (h *CourseHandler) CompleteModule(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	moduleIDStr := c.Param("moduleId")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	err = h.courseService.CompleteModule(courseID, moduleID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module marked as completed"})
}

// LikeCourse godoc
// @Summary      Like a course
// @Description  Adds a like to a course for the authenticated user.
// @Tags         courses
// @Produce      json
// @Param        id  path  string  true  "Course ID (UUID)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/like [post]
func (h *CourseHandler) LikeCourse(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	if err := h.courseService.LikeCourse(courseID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course liked successfully"})
}

// UnlikeCourse godoc
// @Summary      Unlike a course
// @Description  Removes a like from a course for the authenticated user.
// @Tags         courses
// @Produce      json
// @Param        id  path  string  true  "Course ID (UUID)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/like [delete]
func (h *CourseHandler) UnlikeCourse(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	if err := h.courseService.UnlikeCourse(courseID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course unliked successfully"})
}

// GetTrendingCourses godoc
// @Summary      Get trending courses
// @Description  Returns the top trending courses based on enrollment and likes.
// @Tags         courses
// @Produce      json
// @Success      200  {array}   dto.CourseResponse
// @Failure      500  {object}  map[string]string
// @Router       /api/courses/trending [get]
func (h *CourseHandler) GetTrendingCourses(c *gin.Context) {
	var userID uuid.UUID
	if claimsRaw, exists := c.Get(middleware.UserContextKey); exists {
		if claims, ok := claimsRaw.(middleware.UserClaims); ok {
			userID, _ = uuid.Parse(claims.UserID)
		}
	}

	// For now limit to 10 trending courses
	courses, err := h.courseService.GetTrendingCourses(10, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, courses)
}

// AddReview godoc
// @Summary      Add a course review
// @Description  Submits a rating and comment review for a course.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        id    path      string                     true  "Course ID (UUID)"
// @Param        body  body      dto.CreateCourseReviewRequest  true  "Review payload"
// @Success      201   {object}  dto.CourseReviewResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/reviews [post]
func (h *CourseHandler) AddReview(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var req dto.CreateCourseReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.courseService.AddReview(courseID, userID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "course not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetReviews godoc
// @Summary      Get course reviews
// @Description  Returns all reviews for a specific course.
// @Tags         courses
// @Produce      json
// @Param        id  path  string  true  "Course ID (UUID)"
// @Success      200  {array}   dto.CourseReviewResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/courses/{id}/reviews [get]
func (h *CourseHandler) GetReviews(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	reviews, err := h.courseService.GetReviews(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// SearchCourses godoc
// @Summary      Search courses
// @Description  Searches courses by query text with optional filters and pagination.
// @Tags         courses
// @Produce      json
// @Param        query   query  string  false  "Search query text"
// @Param        type    query  string  false  "Course type (free, paid)"
// @Param        format  query  string  false  "Course format (course, project)"
// @Param        status  query  string  false  "Course status (coming soon, active, ended)"
// @Param        limit   query  int     true   "Maximum number of results"
// @Param        offset  query  int     true   "Offset for pagination"
// @Success      200  {array}   dto.CourseSearchResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/courses/search [get]
func (h *CourseHandler) SearchCourses(c *gin.Context) {
	query := c.Query("query")
	courseType := c.Query("type")
	format := c.Query("format")
	status := c.Query("status")
	limit := c.Query("limit")
	offset := c.Query("offset")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
		return
	}
	courses, err := h.courseService.SearchCourses(query, courseType, format, status, limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, courses)
}
