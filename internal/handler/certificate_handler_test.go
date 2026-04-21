package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aruncs31s/gcek_lms_backend/internal/dto"
	"github.com/aruncs31s/gcek_lms_backend/internal/handler"
	"github.com/aruncs31s/gcek_lms_backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MockCertificateService implements service.CertificateService for testing
type MockCertificateService struct {
	GenerateCertificateFunc func(req *dto.GenerateCertificateRequest, baseURL string) (*dto.CertificateResponse, error)
}

func (m *MockCertificateService) GenerateCertificate(req *dto.GenerateCertificateRequest, baseURL string) (*dto.CertificateResponse, error) {
	if m.GenerateCertificateFunc != nil {
		return m.GenerateCertificateFunc(req, baseURL)
	}
	return &dto.CertificateResponse{
		ID:       uuid.New().String(),
		UserID:   req.UserID,
		CourseID: req.CourseID,
		FileURL:  baseURL + "/uploads/certificates/cert_test.pdf",
		IssuedAt: time.Now(),
		Layout:   req.Layout,
	}, nil
}

func setupTestRouter(svc *MockCertificateService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	certHandler := handler.NewCertificateHandler(svc, "http://localhost:8080")

	// Helper middleware to mock authentication context
	authMiddleware := func(c *gin.Context) {
		// Mock a user context to bypass 401 checks in tests that don't specifically test 401s
		c.Set(middleware.UserContextKey, "some-user-id")
		c.Next()
	}

	api := router.Group("/api/certificates")
	api.Use(authMiddleware)
	{
		api.POST("/generate", certHandler.GenerateCertificate)
	}

	// This one doesn't require auth in reality, but just setting it up on router
	router.GET("/api/certificates/download", certHandler.DownloadCertificate)

	return router
}

func TestGenerateCertificate_Unauthorized(t *testing.T) {
	svc := &MockCertificateService{}
	gin.SetMode(gin.TestMode)
	router := gin.New()

	certHandler := handler.NewCertificateHandler(svc, "http://localhost:8080")
	// Notice: NOT using authMiddleware here
	router.POST("/api/certificates/generate", certHandler.GenerateCertificate)

	reqObj := dto.GenerateCertificateRequest{
		UserID:   uuid.NewString(),
		CourseID: uuid.NewString(),
		Layout:   1,
	}
	body, _ := json.Marshal(reqObj)
	req, _ := http.NewRequest(http.MethodPost, "/api/certificates/generate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", w.Code)
	}
}

func TestGenerateCertificate_BadBody(t *testing.T) {
	svc := &MockCertificateService{}
	router := setupTestRouter(svc)

	req, _ := http.NewRequest(http.MethodPost, "/api/certificates/generate", bytes.NewBufferString("{invalid_json}"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", w.Code)
	}
}

func TestGenerateCertificate_ServiceError(t *testing.T) {
	svc := &MockCertificateService{
		GenerateCertificateFunc: func(r *dto.GenerateCertificateRequest, b string) (*dto.CertificateResponse, error) {
			return nil, errors.New("simulated internal error")
		},
	}
	router := setupTestRouter(svc)

	reqObj := dto.GenerateCertificateRequest{
		UserID:   uuid.NewString(),
		CourseID: uuid.NewString(),
		Layout:   5,
	}
	body, _ := json.Marshal(reqObj)
	req, _ := http.NewRequest(http.MethodPost, "/api/certificates/generate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error, got %d", w.Code)
	}
}

func TestGenerateCertificate_Success(t *testing.T) {
	svc := &MockCertificateService{}
	router := setupTestRouter(svc)

	reqObj := dto.GenerateCertificateRequest{
		UserID:   uuid.NewString(),
		CourseID: uuid.NewString(),
		Layout:   7,
	}
	body, _ := json.Marshal(reqObj)
	req, _ := http.NewRequest(http.MethodPost, "/api/certificates/generate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", w.Code)
	}

	var res dto.CertificateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if res.Layout != 7 {
		t.Errorf("Expected layout 7, got %d", res.Layout)
	}
	if res.UserID != reqObj.UserID {
		t.Errorf("Expected user ID %s, got %s", reqObj.UserID, res.UserID)
	}
}

// ─── Download Tests ────────────────────────────────────────────────────────

func TestDownloadCertificate_Validation(t *testing.T) {
	svc := &MockCertificateService{}
	router := setupTestRouter(svc)

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{"Missing File Param", "/api/certificates/download", http.StatusBadRequest},
		{"Empty File Param", "/api/certificates/download?file=", http.StatusBadRequest},
		{"Missing PDF Extension", "/api/certificates/download?file=cert_123", http.StatusBadRequest},
		{"Directory Traversal", "/api/certificates/download?file=../../../etc/passwd.pdf", http.StatusBadRequest},
		{"Not Found", "/api/certificates/download?file=does_not_exist_404.pdf", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestDownloadCertificate_Success(t *testing.T) {
	// Create a dummy PDF file in the actual expected uploads directory relative to working dir
	// However, because paths differ based on test execution context, we need to create it in the root relative path.
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}

	// The handler expects it in "uploads/certificates" relative to the server root (where the app runs).
	// During `go test ./internal/handler/...`, wd is `internal/handler`.
	uploadsDir := filepath.Join(dir, "..", "..", "uploads", "certificates")
	_ = os.MkdirAll(uploadsDir, 0755)

	testFileName := "test_download_cert.pdf"
	testFilePath := filepath.Join(uploadsDir, testFileName)

	// Write dummy content
	_ = os.WriteFile(testFilePath, []byte("%PDF-1.4 mock content"), 0644)
	// Clean up after test
	defer os.Remove(testFilePath)

	svc := &MockCertificateService{}
	router := setupTestRouter(svc)

	url := "/api/certificates/download?file=" + testFileName + "&name=Advanced+Course"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	contentDisposition := w.Header().Get("Content-Disposition")
	expectedDisposition := `attachment; filename="Certificate_Advanced_Course.pdf"`
	if contentDisposition != expectedDisposition {
		t.Errorf("Expected CD %s, got %s", expectedDisposition, contentDisposition)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/pdf" {
		t.Errorf("Expected content type application/pdf, got %s", contentType)
	}
}
