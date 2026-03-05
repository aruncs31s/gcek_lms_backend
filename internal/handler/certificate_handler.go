package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/aruncs/esdc-lms/internal/service"
	"github.com/gin-gonic/gin"
)

type CertificateHandler struct {
	certService service.CertificateService
	baseURL     string
}

func NewCertificateHandler(certService service.CertificateService, baseURL string) *CertificateHandler {
	return &CertificateHandler{
		certService: certService,
		baseURL:     baseURL,
	}
}

func (h *CertificateHandler) GenerateCertificate(c *gin.Context) {
	_, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.GenerateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.certService.GenerateCertificate(&req, h.baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *CertificateHandler) DownloadCertificate(c *gin.Context) {
	fileName := c.Query("file")
	courseName := c.DefaultQuery("name", "Certificate")

	if fileName == "" || len(fileName) > 100 || !strings.HasSuffix(fileName, ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file parameter"})
		return
	}

	// Prevent directory traversal
	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file path"})
		return
	}

	filePath := filepath.Join("uploads/certificates", fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "certificate not found"})
		return
	}

	// Sanitize output filename
	cleanName := strings.ReplaceAll(courseName, " ", "_")
	c.Header("Content-Disposition", "attachment; filename=\"Certificate_"+cleanName+".pdf\"")
	c.Header("Content-Type", "application/pdf")
	c.File(filePath)
}
