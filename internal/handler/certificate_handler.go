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

// GenerateCertificate godoc
// @Summary      Generate a certificate
// @Description  Generates a PDF certificate for a user who completed a course.
// @Tags         certificates
// @Accept       json
// @Produce      json
// @Param        body  body      dto.GenerateCertificateRequest  true  "Certificate generation payload"
// @Success      201   {object}  dto.CertificateResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/certificates/generate [post]
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

// DownloadCertificate godoc
// @Summary      Download a certificate
// @Description  Downloads a certificate PDF file by filename.
// @Tags         certificates
// @Produce      application/pdf
// @Param        file  query  string  true   "Certificate filename (e.g. cert-uuid.pdf)"
// @Param        name  query  string  false  "Course name used in the download filename"
// @Success      200  {file}    binary
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/certificates/download [get]
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
