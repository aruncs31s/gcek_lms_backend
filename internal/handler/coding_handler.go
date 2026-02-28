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
