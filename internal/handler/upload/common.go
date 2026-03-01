package upload

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const (
	ImageUploadType      = "images"
	VideoUploadType      = "videos"
	AttachmentUploadType = "attachments"
)

func getFileURL(
	baseURL,
	fileType,
	fileName string,
) string {
	return fmt.Sprintf("%s/%s/%s", baseURL, fileType, fileName)
}
func getFileName(
	originalName string,
) string {
	ext := filepath.Ext(originalName)
	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
}
