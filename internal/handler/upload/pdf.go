package upload

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/logger"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const PDFUploadType = "pdfs"

type PDFUploadHandler interface {
	Upload(c *gin.Context)
}

type pdfUploadHandler struct {
	UploadDir     string
	BaseUploadURL string
}

func NewPDFUploadHandler(uploadDir, baseURL string) PDFUploadHandler {
	return &pdfUploadHandler{
		UploadDir:     uploadDir,
		BaseUploadURL: baseURL,
	}
}

func (h *pdfUploadHandler) Upload(c *gin.Context) {
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

	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file key, expected 'pdf'"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are allowed"})
		return
	}

	newFileName := getFileName(file.Filename)

	uploadDir := filepath.Join(h.UploadDir, PDFUploadType)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		logger.GetLogger().Error("Failed to create PDF upload directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory"})
		return
	}

	dstPath := filepath.Join(uploadDir, newFileName)
	if err := c.SaveUploadedFile(file, dstPath); err != nil {
		logger.GetLogger().Error("Failed to save PDF file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file"})
		return
	}

	fileURL := getFileURL(h.BaseUploadURL+"/uploads", PDFUploadType, newFileName)
	c.JSON(http.StatusOK, dto.UploadResponse{FileURL: fileURL})
}
