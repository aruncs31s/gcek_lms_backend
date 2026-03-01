package upload

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/logger"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VideoUploadHandler interface {
	Upload(c *gin.Context)
}
type videoUploadHandler struct {
	BaseUploadURL string
}

func NewVideoUploadHandler(baseURL string) *videoUploadHandler {
	return &videoUploadHandler{
		BaseUploadURL: baseURL,
	}
}

func (h *videoUploadHandler) Upload(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)

	if !exists {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "Unauthorized",
			})
		return
	}

	userClaims := userClaimsRaw.(middleware.UserClaims)

	puuid, err := uuid.Parse(userClaims.UserID)

	logger.Log.Debug(
		"Parsed user ID",
		zap.String(
			"userID",
			puuid.String(),
		))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "Invalid file key, expected 'video'",
			})
		return
	}

	contentType := file.Header.Get("Content-Type")
	if contentType != "video/mp4" && contentType != "video/webm" {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "Invalid file type, only mp4 and webm are allowed",
			})
		return
	}

	newFileName := getFileName(
		file.Filename,
	)

	uploadDir := "uploads/videos"
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
		VideoUploadType,
		newFileName,
	)
	c.JSON(http.StatusOK, dto.UploadResponse{FileURL: fileURL})
}
