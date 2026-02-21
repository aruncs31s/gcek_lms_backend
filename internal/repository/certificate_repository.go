package repository

import (
	"github.com/aruncs/esdc-lms/internal/model"
	"gorm.io/gorm"
)

type CertificateRepository interface {
	SaveCertificate(cert *model.Certificate) error
}

type certificateRepository struct {
	db *gorm.DB
}

func NewCertificateRepository(db *gorm.DB) CertificateRepository {
	return &certificateRepository{db}
}

func (r *certificateRepository) SaveCertificate(cert *model.Certificate) error {
	return r.db.Create(cert).Error
}
