package service

type UploadMediaService interface {
	UploadVideo(
		userID string,
		filePath string,
		filename string,
	) (string, error)
	UploadImage(
		userID string,
		filePath string,
		filename string,
	) (string, error)
	UploadAttachment(
		userID string,
		filePath string,
		filename string,
	) (string, error)
	UploadPDF(
		userID string,
		filePath string,
		filename string,
	) (string, error)
}
