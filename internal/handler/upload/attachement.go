package upload

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttachmentUploadHandler interface {
	Upload(c *gin.Context)
}
type attachmentUploadHandler struct {
	BaseUploadURL string
}

func NewAttachmentUploadHandler(baseURL string) AttachmentUploadHandler {
	return &attachmentUploadHandler{
		BaseUploadURL: baseURL,
	}
}

func (h *attachmentUploadHandler) Upload(c *gin.Context) {
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

	file, err := c.FormFile("attachment")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file key, expected 'attachment'"})
		return
	}

	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)

	uploadDir := "uploads/attachments"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory"})
		return
	}

	dstPath := filepath.Join(uploadDir, newFileName)
	if err := c.SaveUploadedFile(file, dstPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file"})
		return
	}

	fileURL := fmt.Sprintf("%s/uploads/attachments/%s", h.BaseUploadURL, newFileName)
	c.JSON(http.StatusOK, dto.UploadResponse{FileURL: fileURL})
}
