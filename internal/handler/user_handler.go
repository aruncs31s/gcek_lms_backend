package handler

import (
	"net/http"
	"strconv"

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
