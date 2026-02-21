package certgen

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

// DocumentModel holds the data needed to populate the certificate HTML template
type DocumentModel struct {
	StudentName string
	CourseName  string
	TeacherName string
	DateIssued  string
}

// Orchestrator handles rendering HTML templates and converting them to PDFs using chromedp Worker
type Orchestrator struct {
	TemplatesDir string
	UploadsDir   string
}

func NewOrchestrator(templatesDir, uploadsDir string) *Orchestrator {
	return &Orchestrator{
		TemplatesDir: templatesDir,
		UploadsDir:   uploadsDir,
	}
}

// GeneratePDF orchestrates the document model injection and PDF printing
func (o *Orchestrator) GeneratePDF(model DocumentModel) (string, error) {
	// 1. Load HTML Template
	tmplPath := filepath.Join(o.TemplatesDir, "certificate.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// 2. Inject DocumentModel into Template
	var renderedHTML bytes.Buffer
	if err := tmpl.Execute(&renderedHTML, model); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Save temporary HTML file
	tempHTMLPath := filepath.Join(os.TempDir(), fmt.Sprintf("cert_%d.html", time.Now().UnixNano()))
	if err := os.WriteFile(tempHTMLPath, renderedHTML.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp html: %w", err)
	}
	defer os.Remove(tempHTMLPath) // clean up

	// 3. Setup Chromedp Worker Context
	// Note: In production you might want an allocator that keeps chrome running a bit longer,
	// but context.WithCancel works well for stateless orchestration.
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Timeout
	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	// 4. Generate PDF bytes using Chromedp worker
	var pdfBuf []byte
	fileURL := "file://" + tempHTMLPath

	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitVisible("#certificate", chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Print to PDF
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithLandscape(true). // A4 landscape mapping
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithPaperWidth(11).   // inches (A4 roughly)
				WithPaperHeight(8.5). // inches (A4 roughly)
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return "", fmt.Errorf("chromedp failed to generate pdf: %w", err)
	}

	// 5. Save the resulting PDF securely to uploads directory
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
