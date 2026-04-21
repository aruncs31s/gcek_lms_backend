package repository

import "github.com/aruncs31s/gcek_lms_backend/pkg/model"

type UploadedMediaRepository interface {
	SaveMedia(
		userID,
		url,
		path,
		filename string,
	) (string, error)
	GetMediaByID(id string) (*model.UploadedMedia, error)
	DeleteMediaByID(id string) error
}
