package handler

import (
	"errors"
	"net/http"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService service.UserService
}

func NewAuthHandler(userService service.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with the given credentials and role.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterRequest  true  "Registration payload"
// @Success      201   {object}  dto.AuthResponse
// @Failure      400   {object}  map[string]string
// @Router       /api/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.userService.RegisterUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// Login godoc
// @Summary      User login
// @Description  Authenticates a user and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Login credentials"
// @Success      200   {object}  dto.AuthResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":   err.Error(),
				"status":  http.StatusBadRequest,
				"success": false,
				"message": "Invalid request payload",
			},
		)
		return
	}

	res, err := h.userService.LoginUser(
		&req,
	)
	if err != nil {
		if errors.Is(service.ErrInvalidCredentials, err) {
			c.JSON(
				http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				},
			)
			return
		}
		c.JSON(
			http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			},
		)
		return
	}

	c.JSON(http.StatusOK, res)
}
