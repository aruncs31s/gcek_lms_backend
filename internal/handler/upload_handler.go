package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadResponse struct {
	FileURL string `json:"file_url"`
}

type UploadHandler struct {
	BaseUploadURL string
}

func NewUploadHandler(baseURL string) *UploadHandler {
	return &UploadHandler{
		BaseUploadURL: baseURL,
	}
}

func (h *UploadHandler) UploadVideo(c *gin.Context) {
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

	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file key, expected 'video'"})
		return
	}

	contentType := file.Header.Get("Content-Type")
	if contentType != "video/mp4" && contentType != "video/webm" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type, only mp4 and webm are allowed"})
		return
	}

	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)

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

	fileURL := fmt.Sprintf("%s/uploads/videos/%s", h.BaseUploadURL, newFileName)
	c.JSON(http.StatusOK, UploadResponse{FileURL: fileURL})
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
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

	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)

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

	fileURL := fmt.Sprintf("%s/uploads/images/%s", h.BaseUploadURL, newFileName)
	c.JSON(http.StatusOK, UploadResponse{FileURL: fileURL})
}

func (h *UploadHandler) UploadAttachment(c *gin.Context) {
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
	c.JSON(http.StatusOK, UploadResponse{FileURL: fileURL})
}
