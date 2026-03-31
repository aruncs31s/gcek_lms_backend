package certgen

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/aruncs/esdc-lms/utils"
	"github.com/google/uuid"
)

// DocumentModel holds the data needed to populate the certificate HTML template
type DocumentModel struct {
	StudentName string
	CourseName  string
	TeacherName string
	DateIssued  string
	// Layout selects which of the 10 templates to use (1-10). Defaults to 1.
	Layout int
	// LogoDataURI is the base64-encoded PNG logo embedded as a data URI
	LogoDataURI template.URL
}

// Orchestrator defines the interface for generating PDF certificates.
type Orchestrator interface {
	GeneratePDF(model DocumentModel) (string, error)
}

// orchestratorImpl handles rendering HTML templates and converting them to PDFs using Chrome Pool.
type orchestratorImpl struct {
	TemplatesDir string
	UploadsDir   string
}

func NewOrchestrator(templatesDir, uploadsDir string) Orchestrator {
	return &orchestratorImpl{
		TemplatesDir: templatesDir,
		UploadsDir:   uploadsDir,
	}
}

// loadLogoAsDataURI reads the logo PNG and returns a base64 data URI string.
func (o *orchestratorImpl) loadLogoAsDataURI() template.URL {
	logoPath := filepath.Join(o.TemplatesDir, "logo.png")
	data, err := os.ReadFile(logoPath)
	if err != nil {
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return template.URL("data:image/png;base64," + encoded)
}

// GeneratePDF orchestrates the document model injection and PDF printing
func (o *orchestratorImpl) GeneratePDF(model DocumentModel) (string, error) {
	// Validate layout range
	layout := model.Layout
	if layout < 1 || layout > 10 {
		layout = 1
	}

	// Inject logo as data URI so the HTML file loaded from /tmp can render it
	model.LogoDataURI = o.loadLogoAsDataURI()

	// 1. Load HTML Template based on layout number
	templateName := fmt.Sprintf("layout_%d.html", layout)
	tmplPath := filepath.Join(o.TemplatesDir, templateName)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// 2. Inject DocumentModel into Template
	var renderedHTML bytes.Buffer
	if err := tmpl.Execute(&renderedHTML, model); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// 3. Convert HTML string to PDF using pooled Chrome instance
	opts := &utils.PDFOptions{
		Landscape:       false, // Setting false here, since Width > Height automatically creates a landscape page
		PrintBackground: true,
		PaperWidth:      11,
		PaperHeight:     8.5,
		MarginTop:       0,
		MarginBottom:    0,
		MarginLeft:      0,
		MarginRight:     0,
	}

	pdfBuf, err := utils.ConvertHTMLToPDF(context.Background(), renderedHTML.String(), opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate pdf: %w", err)
	}

	// 4. Save the resulting PDF securely to uploads directory
	if err := os.MkdirAll(o.UploadsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create uploads dir: %w", err)
	}

	fileName := fmt.Sprintf("cert_%d_%s.pdf", time.Now().Unix(), uuid.New().String()[:6])
	finalPath := filepath.Join(o.UploadsDir, fileName)

	if err := os.WriteFile(finalPath, pdfBuf, 0644); err != nil {
		return "", fmt.Errorf("failed to save pdf: %w", err)
	}

	return fileName, nil
}
