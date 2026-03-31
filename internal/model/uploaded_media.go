package model

// TODO: Implement..
const ()

type UploadedMedia struct {
	ID       string `gorm:"type:varchar(255);primaryKey"`
	UserID   string `gorm:"type:varchar(255);not null"`
	URL      string `gorm:"type:varchar(255);not null"`
	Path     string `gorm:"type:varchar(255);not null"`
	Filename string `gorm:"type:varchar(255);not null"`
	Type     string `gorm:"type:varchar(50);not null"`
}

func (u *UploadedMedia) TableName() string {
	return "uploaded_media"
}
