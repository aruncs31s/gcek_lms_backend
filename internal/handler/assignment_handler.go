package handler

import (
	"net/http"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/aruncs/esdc-lms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AssignmentHandler struct {
	assignmentService service.AssignmentService
}

func NewAssignmentHandler(assignmentService service.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{assignmentService: assignmentService}
}

// ==========================================
// Handlers
// ==========================================

// CreateAssignment godoc
// @Summary      Create an assignment
// @Description  Creates a new assignment for a course. Requires Teacher or Admin role.
// @Tags         assignments
// @Accept       json
// @Produce      json
// @Param        id    path      string                      true  "Course ID (UUID)"
// @Param        body  body      dto.CreateAssignmentRequest true  "Assignment creation payload"
// @Success      201   {object}  dto.AssignmentResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/assignments [post]
func (h *AssignmentHandler) CreateAssignment(c *gin.Context) {
	teacherID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	courseID, err := getCourseIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var req dto.CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.assignmentService.CreateAssignment(courseID, teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetAssignments godoc
// @Summary      List assignments
// @Description  Returns all assignments for a course.
// @Tags         assignments
// @Produce      json
// @Param        id  path  string  true  "Course ID (UUID)"
// @Success      200  {array}   dto.AssignmentResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/assignments [get]
func (h *AssignmentHandler) GetAssignments(c *gin.Context) {
	userID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	courseID, err := getCourseIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	res, err := h.assignmentService.GetAssignmentsByCourse(courseID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetAssignmentByID godoc
// @Summary      Get assignment by ID
// @Description  Returns details of a specific assignment.
// @Tags         assignments
// @Produce      json
// @Param        id            path  string  true  "Course ID (UUID)"
// @Param        assignmentId  path  string  true  "Assignment ID (UUID)"
// @Success      200  {object}  dto.AssignmentResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/assignments/{assignmentId} [get]
func (h *AssignmentHandler) GetAssignmentByID(c *gin.Context) {
	userID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	courseID, err := getCourseIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	assignmentIDStr := c.Param("assignmentId")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	res, err := h.assignmentService.GetAssignmentByID(courseID, assignmentID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// UpdateAssignment godoc
// @Summary      Update an assignment
// @Description  Updates an existing assignment. Requires Teacher or Admin role.
// @Tags         assignments
// @Accept       json
// @Produce      json
// @Param        id            path      string                      true  "Course ID (UUID)"
// @Param        assignmentId  path      string                      true  "Assignment ID (UUID)"
// @Param        body          body      dto.UpdateAssignmentRequest true  "Assignment update payload"
// @Success      200   {object}  dto.AssignmentResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/assignments/{assignmentId} [put]
func (h *AssignmentHandler) UpdateAssignment(c *gin.Context) {
	teacherID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	courseID, err := getCourseIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	assignmentIDStr := c.Param("assignmentId")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	var req dto.UpdateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.assignmentService.UpdateAssignment(courseID, assignmentID, teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// DeleteAssignment godoc
// @Summary      Delete an assignment
// @Description  Deletes an assignment from a course. Requires Teacher or Admin role.
// @Tags         assignments
// @Produce      json
// @Param        id            path  string  true  "Course ID (UUID)"
// @Param        assignmentId  path  string  true  "Assignment ID (UUID)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/assignments/{assignmentId} [delete]
func (h *AssignmentHandler) DeleteAssignment(c *gin.Context) {
	teacherID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	courseID, err := getCourseIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	assignmentIDStr := c.Param("assignmentId")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	err = h.assignmentService.DeleteAssignment(courseID, assignmentID, teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment deleted successfully"})
}

// SubmitAssignment godoc
// @Summary      Submit an assignment
// @Description  Submits a file URL as the student's assignment submission.
// @Tags         assignments
// @Accept       json
// @Produce      json
// @Param        id            path      string                       true  "Course ID (UUID)"
// @Param        assignmentId  path      string                       true  "Assignment ID (UUID)"
// @Param        body          body      dto.SubmitAssignmentRequest  true  "Submission payload"
// @Success      201   {object}  dto.AssignmentSubmissionResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/assignments/{assignmentId}/submit [post]
func (h *AssignmentHandler) SubmitAssignment(c *gin.Context) {
	studentID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	courseID, err := getCourseIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	assignmentIDStr := c.Param("assignmentId")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	var req dto.SubmitAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.assignmentService.SubmitAssignment(courseID, assignmentID, studentID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetSubmissions godoc
// @Summary      Get all submissions
// @Description  Returns all student submissions for an assignment. Requires Teacher or Admin role.
// @Tags         assignments
// @Produce      json
// @Param        id            path  string  true  "Course ID (UUID)"
// @Param        assignmentId  path  string  true  "Assignment ID (UUID)"
// @Success      200  {array}   dto.AssignmentSubmissionResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/assignments/{assignmentId}/submissions [get]
func (h *AssignmentHandler) GetSubmissions(c *gin.Context) {
	teacherID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	courseID, err := getCourseIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	assignmentIDStr := c.Param("assignmentId")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	res, err := h.assignmentService.GetSubmissions(courseID, assignmentID, teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GradeSubmission godoc
// @Summary      Grade a submission
// @Description  Assigns a score and feedback to a student's submission. Requires Teacher or Admin role.
// @Tags         assignments
// @Accept       json
// @Produce      json
// @Param        id            path      string                    true  "Course ID (UUID)"
// @Param        assignmentId  path      string                    true  "Assignment ID (UUID)"
// @Param        submissionId  path      string                    true  "Submission ID (UUID)"
// @Param        body          body      dto.GradeSubmissionRequest true  "Grading payload"
// @Success      200   {object}  dto.AssignmentSubmissionResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/assignments/{assignmentId}/submissions/{submissionId}/grade [put]
func (h *AssignmentHandler) GradeSubmission(c *gin.Context) {
	teacherID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	courseID, err := getCourseIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	assignmentIDStr := c.Param("assignmentId")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	submissionIDStr := c.Param("submissionId")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	var req dto.GradeSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.assignmentService.GradeSubmission(courseID, assignmentID, submissionID, teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetStudentSubmission godoc
// @Summary      Get own submission
// @Description  Returns the authenticated student's submission for a specific assignment.
// @Tags         assignments
// @Produce      json
// @Param        id            path  string  true  "Course ID (UUID)"
// @Param        assignmentId  path  string  true  "Assignment ID (UUID)"
// @Success      200  {object}  dto.AssignmentSubmissionResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/assignments/{assignmentId}/submissions/me [get]
func (h *AssignmentHandler) GetStudentSubmission(c *gin.Context) {
	studentID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	courseID, err := getCourseIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	assignmentIDStr := c.Param("assignmentId")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	res, err := h.assignmentService.GetStudentSubmission(courseID, assignmentID, studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if res == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// ==========================================
// Helpers
// ==========================================

func getUserIdFromContext(c *gin.Context) (uuid.UUID, error) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		return uuid.Nil, http.ErrNoCookie
	}
	userClaims := userClaimsRaw.(middleware.UserClaims)
	return uuid.Parse(userClaims.UserID)
}

func getCourseIdFromContext(c *gin.Context) (uuid.UUID, error) {
	idStr := c.Param("id")
	return uuid.Parse(idStr)
}
