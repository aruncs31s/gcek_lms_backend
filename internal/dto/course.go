package dto

import "time"

type CreateCourseRequest struct {
	Title        string  `json:"title" validate:"required"`
	Description  string  `json:"description"`
	ThumbnailURL string  `json:"thumbnail_url"`
	Price        float64 `json:"price" validate:"gte=0"`
}

type UpdateCourseRequest struct {
	Title        *string  `json:"title"`
	Description  *string  `json:"description"`
	ThumbnailURL *string  `json:"thumbnail_url"`
	Price        *float64 `json:"price" validate:"omitempty,gte=0"`
}

type CourseResponse struct {
	ID           string           `json:"id"`
	TeacherID    string           `json:"teacher_id"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	ThumbnailURL string           `json:"thumbnail_url"`
	Price        float64          `json:"price"`
	CreatedAt    time.Time        `json:"created_at"`
	Modules      []ModuleResponse `json:"modules,omitempty"`
}

type CreateModuleRequest struct {
	Title    string `json:"title" validate:"required"`
	VideoURL string `json:"video_url"`
}

type ModuleResponse struct {
	ID         string `json:"id"`
	CourseID   string `json:"course_id"`
	Title      string `json:"title"`
	VideoURL   string `json:"video_url"`
	OrderIndex int    `json:"order_index"`
}
