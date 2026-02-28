package dto

import "time"

type CreateCourseRequest struct {
	Title        string  `json:"title" validate:"required"`
	Description  string  `json:"description"`
	ThumbnailURL string  `json:"thumbnail_url"`
	Price        float64 `json:"price" validate:"gte=0"`
	Type         string  `json:"type" validate:"omitempty,oneof=free paid"`
	Status       string  `json:"status" validate:"omitempty,oneof='not started' active ended"`
}

type UpdateCourseRequest struct {
	Title        *string  `json:"title"`
	Description  *string  `json:"description"`
	ThumbnailURL *string  `json:"thumbnail_url"`
	Price        *float64 `json:"price" validate:"omitempty,gte=0"`
	Type         *string  `json:"type" validate:"omitempty,oneof=free paid"`
	Status       *string  `json:"status" validate:"omitempty,oneof='not started' active ended"`
}

type CourseResponse struct {
	ID            string           `json:"id"`
	TeacherID     string           `json:"teacher_id"`
	TeacherName   string           `json:"teacher_name,omitempty"`
	TeacherAvatar string           `json:"teacher_avatar_url,omitempty"`
	TeacherBio    string           `json:"teacher_bio,omitempty"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	ThumbnailURL  string           `json:"thumbnail_url"`
	Price         float64          `json:"price"`
	Type          string           `json:"type"`
	Status        string           `json:"status"`
	CreatedAt     time.Time        `json:"created_at"`
	StudentCount  int              `json:"student_count"`
	Modules       []ModuleResponse `json:"modules,omitempty"`
}

type CreateModuleRequest struct {
	ParentID    *string `json:"parent_id"`
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description"`
	Type        string  `json:"type" validate:"required"` // video or chapter
	VideoURL    string  `json:"video_url"`
	Points      int     `json:"points"`
	IsFree      bool    `json:"is_free"`
}

type UpdateModuleRequest struct {
	ParentID    *string `json:"parent_id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	VideoURL    *string `json:"video_url"`
	Points      *int    `json:"points"`
	IsFree      *bool   `json:"is_free"`
}

type ReorderModulesRequest struct {
	ModuleIDs []string `json:"module_ids" validate:"required"`
}

type ModuleResponse struct {
	ID          string  `json:"id"`
	CourseID    string  `json:"course_id"`
	ParentID    *string `json:"parent_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	VideoURL    string  `json:"video_url"`
	Points      int     `json:"points"`
	IsFree      bool    `json:"is_free"`
	OrderIndex  int     `json:"order_index"`
	IsCompleted bool    `json:"is_completed"`
}
