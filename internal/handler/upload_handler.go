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
func NewUploadHandler(uploadDir, baseURL string) UploadHandler {

	video := upload.NewVideoUploadHandler(uploadDir, baseURL)
	image := upload.NewImageUploadHandler(uploadDir, baseURL)
	attachement := upload.NewAttachmentUploadHandler(uploadDir, baseURL)
	uploadHandler := &uploadHandler{
		BaseUploadURL: baseURL,
		uploadType:    make(map[string]UploadWriter),
	}

	uploadHandler.provide("video", video)
	uploadHandler.provide("image", image)
	uploadHandler.provide("attachment", attachement)
	return uploadHandler
}
// UploadVideo godoc
// @Summary      Upload a video
// @Description  Uploads a video file (mp4 or webm). Requires Teacher or Admin role.
// @Tags         uploads
// @Accept       multipart/form-data
// @Produce      json
// @Param        video  formData  file  true  "Video file (mp4 or webm)"
// @Success      200  {object}  dto.UploadResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/upload/video [post]
func (h *uploadHandler) UploadVideo(c *gin.Context) {
	h.uploadType["video"].Upload(c)
}
// UploadImage godoc
// @Summary      Upload an image
// @Description  Uploads an image file (jpg, jpeg, png, gif, webp).
// @Tags         uploads
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "Image file"
// @Success      200  {object}  dto.UploadResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/upload/image [post]
func (h *uploadHandler) UploadImage(c *gin.Context) {
	h.uploadType["image"].Upload(c)
}

// UploadAttachment godoc
// @Summary      Upload an attachment
// @Description  Uploads an attachment file (pdf, doc, docx, txt).
// @Tags         uploads
// @Accept       multipart/form-data
// @Produce      json
// @Param        attachment  formData  file  true  "Attachment file"
// @Success      200  {object}  dto.UploadResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/upload/attachment [post]
func (h *uploadHandler) UploadAttachment(c *gin.Context) {
	h.uploadType["attachment"].Upload(c)
}
