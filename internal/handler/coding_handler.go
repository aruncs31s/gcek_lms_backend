package handler

import (
	"net/http"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/aruncs/esdc-lms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CodingHandler struct {
	codingService service.CodingService
}

func NewCodingHandler(codingService service.CodingService) *CodingHandler {
	return &CodingHandler{codingService: codingService}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper
// ─────────────────────────────────────────────────────────────────────────────

func getCodingAssignmentIDFromContext(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.Param("codingAssignmentId"))
}

func getCodingSubmissionIDFromContext(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.Param("submissionId"))
}

func isTeacherRole(c *gin.Context) bool {
	raw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		return false
	}
	claims, ok := raw.(middleware.UserClaims)
	if !ok {
		return false
	}
	return claims.Role == "teacher" || claims.Role == "admin"
}

// ─────────────────────────────────────────────────────────────────────────────
// Teacher: CRUD
// ─────────────────────────────────────────────────────────────────────────────

// CreateCodingAssignment godoc
// @Summary      Create a coding assignment
// @Description  Creates a new coding assignment with test cases for a course. Requires Teacher or Admin role.
// @Tags         coding-assignments
// @Accept       json
// @Produce      json
// @Param        id    path      string                              true  "Course ID (UUID)"
// @Param        body  body      dto.CreateCodingAssignmentRequest   true  "Coding assignment creation payload"
// @Success      201   {object}  dto.CodingAssignmentResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/coding-assignments [post]
// POST /api/courses/:id/coding-assignments
func (h *CodingHandler) CreateCodingAssignment(c *gin.Context) {
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

	var req dto.CreateCodingAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.codingService.CreateCodingAssignment(courseID, teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// UpdateCodingAssignment godoc
// @Summary      Update a coding assignment
// @Description  Updates an existing coding assignment. Requires Teacher or Admin role.
// @Tags         coding-assignments
// @Accept       json
// @Produce      json
// @Param        id                  path      string                             true  "Course ID (UUID)"
// @Param        codingAssignmentId  path      string                             true  "Coding Assignment ID (UUID)"
// @Param        body                body      dto.UpdateCodingAssignmentRequest  true  "Update payload"
// @Success      200   {object}  dto.CodingAssignmentResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/coding-assignments/{codingAssignmentId} [put]
// PUT /api/courses/:id/coding-assignments/:codingAssignmentId
func (h *CodingHandler) UpdateCodingAssignment(c *gin.Context) {
	teacherID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	assignmentID, err := getCodingAssignmentIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coding assignment ID"})
		return
	}

	var req dto.UpdateCodingAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.codingService.UpdateCodingAssignment(assignmentID, teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// DeleteCodingAssignment godoc
// @Summary      Delete a coding assignment
// @Description  Deletes a coding assignment. Requires Teacher or Admin role.
// @Tags         coding-assignments
// @Produce      json
// @Param        id                  path  string  true  "Course ID (UUID)"
// @Param        codingAssignmentId  path  string  true  "Coding Assignment ID (UUID)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/coding-assignments/{codingAssignmentId} [delete]
// DELETE /api/courses/:id/coding-assignments/:codingAssignmentId
func (h *CodingHandler) DeleteCodingAssignment(c *gin.Context) {
	teacherID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	assignmentID, err := getCodingAssignmentIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coding assignment ID"})
		return
	}

	if err := h.codingService.DeleteCodingAssignment(assignmentID, teacherID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coding assignment deleted"})
}

// GetCodingSubmissions godoc
// @Summary      Get all coding submissions
// @Description  Returns all student submissions for a coding assignment. Requires Teacher or Admin role.
// @Tags         coding-assignments
// @Produce      json
// @Param        id                  path  string  true  "Course ID (UUID)"
// @Param        codingAssignmentId  path  string  true  "Coding Assignment ID (UUID)"
// @Success      200  {array}   dto.CodingSubmissionResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/coding-assignments/{codingAssignmentId}/submissions [get]
// GET /api/courses/:id/coding-assignments/:codingAssignmentId/submissions
func (h *CodingHandler) GetSubmissions(c *gin.Context) {
	teacherID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	assignmentID, err := getCodingAssignmentIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coding assignment ID"})
		return
	}

	subs, err := h.codingService.GetSubmissions(assignmentID, teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subs)
}

// GradeCodingSubmission godoc
// @Summary      Grade a coding submission
// @Description  Assigns a score and feedback to a student's coding submission. Requires Teacher or Admin role.
// @Tags         coding-assignments
// @Accept       json
// @Produce      json
// @Param        id                  path      string                           true  "Course ID (UUID)"
// @Param        codingAssignmentId  path      string                           true  "Coding Assignment ID (UUID)"
// @Param        submissionId        path      string                           true  "Submission ID (UUID)"
// @Param        body                body      dto.GradeCodingSubmissionRequest  true  "Grading payload"
// @Success      200   {object}  dto.CodingSubmissionResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/coding-assignments/{codingAssignmentId}/submissions/{submissionId}/grade [put]
// PUT /api/courses/:id/coding-assignments/:codingAssignmentId/submissions/:submissionId/grade
func (h *CodingHandler) GradeSubmission(c *gin.Context) {
	teacherID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	submissionID, err := getCodingSubmissionIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	var req dto.GradeCodingSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.codingService.GradeSubmission(submissionID, teacherID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// ─────────────────────────────────────────────────────────────────────────────
// Shared: list & detail
// ─────────────────────────────────────────────────────────────────────────────

// GetCodingAssignments godoc
// @Summary      List coding assignments
// @Description  Returns all coding assignments for a course.
// @Tags         coding-assignments
// @Produce      json
// @Param        id  path  string  true  "Course ID (UUID)"
// @Success      200  {array}   dto.CodingAssignmentResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/coding-assignments [get]
// GET /api/courses/:id/coding-assignments
func (h *CodingHandler) GetCodingAssignments(c *gin.Context) {
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

	res, err := h.codingService.GetCodingAssignmentsByCourse(courseID, userID, isTeacherRole(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetCodingAssignmentByID godoc
// @Summary      Get coding assignment by ID
// @Description  Returns details of a specific coding assignment including test cases.
// @Tags         coding-assignments
// @Produce      json
// @Param        id                  path  string  true  "Course ID (UUID)"
// @Param        codingAssignmentId  path  string  true  "Coding Assignment ID (UUID)"
// @Success      200  {object}  dto.CodingAssignmentResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/coding-assignments/{codingAssignmentId} [get]
// GET /api/courses/:id/coding-assignments/:codingAssignmentId
func (h *CodingHandler) GetCodingAssignmentByID(c *gin.Context) {
	userID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	assignmentID, err := getCodingAssignmentIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coding assignment ID"})
		return
	}

	res, err := h.codingService.GetCodingAssignmentByID(assignmentID, userID, isTeacherRole(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// ─────────────────────────────────────────────────────────────────────────────
// Student: run & submit
// ─────────────────────────────────────────────────────────────────────────────

// RunCode godoc
// @Summary      Run code in sandbox
// @Description  Executes code in a sandboxed environment without submitting it as an assignment.
// @Tags         coding-assignments
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RunCodeRequest  true  "Code execution payload"
// @Success      200   {object}  dto.RunCodeResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/code/run [post]
// POST /api/code/run   (no assignment context – free execution sandbox)
func (h *CodingHandler) RunCode(c *gin.Context) {
	var req dto.RunCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.codingService.RunCode(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// SubmitCode godoc
// @Summary      Submit code for a coding assignment
// @Description  Submits code for evaluation against the test cases of a coding assignment.
// @Tags         coding-assignments
// @Accept       json
// @Produce      json
// @Param        id                  path      string                  true  "Course ID (UUID)"
// @Param        codingAssignmentId  path      string                  true  "Coding Assignment ID (UUID)"
// @Param        body                body      dto.SubmitCodingRequest true  "Code submission payload"
// @Success      200   {object}  dto.CodingSubmissionResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/coding-assignments/{codingAssignmentId}/submit [post]
// POST /api/courses/:id/coding-assignments/:codingAssignmentId/submit
func (h *CodingHandler) SubmitCode(c *gin.Context) {
	studentID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	assignmentID, err := getCodingAssignmentIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coding assignment ID"})
		return
	}

	var req dto.SubmitCodingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.codingService.SubmitCode(assignmentID, studentID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetMySubmission godoc
// @Summary      Get own coding submission
// @Description  Returns the authenticated student's submission for a specific coding assignment.
// @Tags         coding-assignments
// @Produce      json
// @Param        id                  path  string  true  "Course ID (UUID)"
// @Param        codingAssignmentId  path  string  true  "Coding Assignment ID (UUID)"
// @Success      200  {object}  dto.CodingSubmissionResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/courses/{id}/coding-assignments/{codingAssignmentId}/submissions/me [get]
// GET /api/courses/:id/coding-assignments/:codingAssignmentId/submissions/me
func (h *CodingHandler) GetMySubmission(c *gin.Context) {
	studentID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	assignmentID, err := getCodingAssignmentIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coding assignment ID"})
		return
	}

	res, err := h.codingService.GetMySubmission(assignmentID, studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, res)
}
