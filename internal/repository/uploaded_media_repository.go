package repository

import "github.com/aruncs/esdc-lms/internal/model"

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
