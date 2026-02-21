package handler

import (
	"net/http"

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
