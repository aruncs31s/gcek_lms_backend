package handler

import (
	"github.com/aruncs/esdc-lms/internal/handler/upload"
	"github.com/gin-gonic/gin"
)

type UploadWriter interface {
	Upload(c *gin.Context)
}
type UploadHandler interface {
	UploadVideo(c *gin.Context)
	UploadImage(c *gin.Context)
	UploadAttachment(c *gin.Context)
}

type uploadHandler struct {
	BaseUploadURL string
	uploadType    map[string]UploadWriter
}

func (h *uploadHandler) provide(
	uploadType string,
	handler UploadWriter,
) {
	h.uploadType[uploadType] = handler
}
func NewUploadHandler(baseURL string) UploadHandler {

	video := upload.NewVideoUploadHandler(baseURL)
	image := upload.NewImageUploadHandler(baseURL)
	attachement := upload.NewAttachmentUploadHandler(baseURL)
	uploadHandler := &uploadHandler{
		BaseUploadURL: baseURL,
		uploadType:    make(map[string]UploadWriter),
	}

	uploadHandler.provide("video", video)
	uploadHandler.provide("image", image)
	uploadHandler.provide("attachment", attachement)
	return uploadHandler
}
func (h *uploadHandler) UploadVideo(c *gin.Context) {
	h.uploadType["video"].Upload(c)
}
func (h *uploadHandler) UploadImage(c *gin.Context) {
	h.uploadType["image"].Upload(c)
}

func (h *uploadHandler) UploadAttachment(c *gin.Context) {
	h.uploadType["attachment"].Upload(c)
}
