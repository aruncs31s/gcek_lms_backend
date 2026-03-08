package handler

import (
	"net/http"
	"strconv"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/aruncs/esdc-lms/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	us service.UserService
}

func NewUserHandler(us service.UserService) *UserHandler {
	return &UserHandler{us: us}
}

func (h *UserHandler) List(
	c *gin.Context,
) {
	limit := c.Query("limit")
	offset := c.Query("offset")

	userType := c.Query("user_type")
	// Parse To Int
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 50
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		offsetInt = 0
	}
	users, count, err := h.us.List(limitInt, offsetInt, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "total_users": count})
}

func (h *UserHandler) Enrolments(c *gin.Context) {
	userId := c.Param("id")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	res, err := h.us.Enrolments(limit, offset, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userClaims, ok := userClaimsRaw.(middleware.UserClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.us.UpdateProfile(userClaims.UserID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *UserHandler) Search(c *gin.Context) {
	query := c.Query("query")

	role := c.Query("role")

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, count, err := h.us.Search(query, role, limit, offset)
	if err != nil {
		users = []dto.UserResponse{}
		count = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"total_users": count,
	})
}
