package upload

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageUploadHandler interface {
	Upload(c *gin.Context)
}
type imageUploadHandler struct {
	BaseUploadURL string
}

func NewImageUploadHandler(baseURL string) ImageUploadHandler {
	return &imageUploadHandler{
		BaseUploadURL: baseURL,
	}
}

func (h *imageUploadHandler) Upload(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userClaims := userClaimsRaw.(middleware.UserClaims)
	_, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file key, expected 'image'"})
		return
	}

	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type, only jpeg, png and webp are allowed"})
		return
	}

	// ext := filepath.Ext(file.Filename)
	newFileName := getFileName(file.Filename)

	uploadDir := "uploads/images"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory"})
		return
	}

	dstPath := filepath.Join(uploadDir, newFileName)
	if err := c.SaveUploadedFile(file, dstPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file"})
		return
	}

	fileURL := getFileURL(
		h.BaseUploadURL,
		ImageUploadType,
		newFileName,
	)
	c.JSON(http.StatusOK, dto.UploadResponse{FileURL: fileURL})
}
