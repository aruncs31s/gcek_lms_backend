package certgen

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLoadLogoAsDataURI(t *testing.T) {
	// Use the real templates directory
	templatesDir := filepath.Join("..", "..", "templates")

	orc := NewOrchestrator(templatesDir, os.TempDir())
	uri := orc.loadLogoAsDataURI()
	uriStr := string(uri)

	if !strings.HasPrefix(uriStr, "data:image/png;base64,") {
		t.Errorf("Expected data URI to start with data:image/png;base64, got %s", uriStr[:min(len(uriStr), 30)])
	}
	if len(uriStr) < 100 {
		t.Errorf("Data URI seems too short to be a valid image: length %d", len(uriStr))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestGeneratePDF_DefaultLayout(t *testing.T) {
	templatesDir := filepath.Join("..", "..", "templates")
	outDir := filepath.Join("..", "..", "testdata", "output")
	_ = os.MkdirAll(outDir, 0755)

	orc := NewOrchestrator(templatesDir, outDir)

	model := DocumentModel{
		StudentName: "Test Student Default",
		CourseName:  "Test Course Basics",
		TeacherName: "Prof. Default",
		DateIssued:  "Jan 01, 2026",
		Layout:      0, // Should default to 1
	}

	fileName, err := orc.GeneratePDF(model)
	if err != nil {
		t.Fatalf("GeneratePDF failed: %v", err)
	}

	if fileName == "" {
		t.Fatal("Expected a file name, got empty string")
	}

	if !strings.HasSuffix(fileName, ".pdf") {
		t.Errorf("Expected filename ending in .pdf, got %s", fileName)
	}

	// Verify file exists
	path := filepath.Join(outDir, fileName)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Expected PDF file to exist at %s, got error: %v", path, err)
	}
	if info.Size() == 0 {
		t.Errorf("Expected PDF file to not be empty")
	}

	// Clean up this test's specific file since it has a random UUID
	_ = os.Remove(path)
}

// TestGeneratePDF_AllLayouts runs through layouts 1-10 sequentially to avoid Chrome timeout issues.
func TestGeneratePDF_AllLayouts(t *testing.T) {
	templatesDir := filepath.Join("..", "..", "templates")
	outDir := filepath.Join("..", "..", "testdata", "output")
	_ = os.MkdirAll(outDir, 0755)

	orc := NewOrchestrator(templatesDir, outDir)

	for layout := 1; layout <= 10; layout++ {
		t.Run("Layout_"+strconv.Itoa(layout), func(t *testing.T) {
			model := DocumentModel{
				StudentName: "Test Student " + strconv.Itoa(layout),
				CourseName:  "Course " + strconv.Itoa(layout),
				TeacherName: "Prof. " + strconv.Itoa(layout),
				DateIssued:  "Jan 01, 2026",
				Layout:      layout,
			}

			fileName, err := orc.GeneratePDF(model)
			if err != nil {
				t.Fatalf("GeneratePDF failed for layout %d: %v", layout, err)
			}

			path := filepath.Join(outDir, fileName)
			// Rename it to a predictable name so we can view them easily
			targetPath := filepath.Join(outDir, "test_generate_layout_"+strconv.Itoa(layout)+".pdf")
			_ = os.Rename(path, targetPath)

			info, err := os.Stat(targetPath)
			if err != nil {
				t.Fatalf("Expected PDF file to exist at %s: %v", targetPath, err)
			}
			if info.Size() == 0 {
				t.Errorf("Expected PDF file to not be empty")
			}
			// Let these stay in testdata/output/ so user can view them
		})
	}
}

func TestGeneratePDF_InvalidTemplateDir(t *testing.T) {
	// Point to a directory that definitely does not have layout_1.html
	orc := NewOrchestrator(os.TempDir(), os.TempDir())

	model := DocumentModel{
		StudentName: "Test Student",
		Layout:      1,
	}

	_, err := orc.GeneratePDF(model)
	if err == nil {
		t.Fatal("Expected error when using invalid template directory, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse template") {
		t.Errorf("Expected template parse error, got: %v", err)
	}
}
