package dto

import "time"

type CreateCourseRequest struct {
	Title                  string     `json:"title" validate:"required"`
	Description            string     `json:"description"`
	ThumbnailURL           string     `json:"thumbnail_url"`
	Price                  float64    `json:"price" validate:"gte=0"`
	Type                   string     `json:"type" validate:"omitempty,oneof=free paid"`
	Format                 string     `json:"format" validate:"omitempty,oneof=course project"`
	Status                 string     `json:"status" validate:"omitempty,oneof='coming soon' active ended"`
	Duration               string     `json:"duration"`
	IsCertificateAvailable bool       `json:"is_certificate_available"`
	StartDate              *time.Time `json:"start_date"`
}

type UpdateCourseRequest struct {
	Title                  *string    `json:"title"`
	Description            *string    `json:"description"`
	ThumbnailURL           *string    `json:"thumbnail_url"`
	Price                  *float64   `json:"price" validate:"omitempty,gte=0"`
	Type                   *string    `json:"type" validate:"omitempty,oneof=free paid"`
	Format                 *string    `json:"format" validate:"omitempty,oneof=course project"`
	Status                 *string    `json:"status" validate:"omitempty,oneof='coming soon' active ended"`
	Duration               *string    `json:"duration"`
	IsCertificateAvailable *bool      `json:"is_certificate_available"`
	StartDate              *time.Time `json:"start_date"`
}

type CourseResponse struct {
	ID                     string           `json:"id"`
	TeacherID              string           `json:"teacher_id"`
	TeacherName            string           `json:"teacher_name,omitempty"`
	TeacherAvatar          string           `json:"teacher_avatar_url,omitempty"`
	TeacherBio             string           `json:"teacher_bio,omitempty"`
	Title                  string           `json:"title"`
	Description            string           `json:"description"`
	ThumbnailURL           string           `json:"thumbnail_url"`
	Price                  float64          `json:"price"`
	Type                   string           `json:"type"`
	Format                 string           `json:"format"`
	Status                 string           `json:"status"`
	Duration               string           `json:"duration,omitempty"`
	IsCertificateAvailable bool             `json:"is_certificate_available"`
	StartDate              *time.Time       `json:"start_date,omitempty"`
	Progress               float64          `json:"progress,omitempty"`
	CreatedAt              time.Time        `json:"created_at"`
	StudentCount           int              `json:"student_count"`
	LikesCount             int64            `json:"likes_count"`
	IsLiked                bool             `json:"is_liked"`
	Modules                []ModuleResponse `json:"modules,omitempty"`
}

type CourseSearchResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
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

type CreateCourseReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"required"`
}

type CourseReviewResponse struct {
	ID        string    `json:"id"`
	CourseID  string    `json:"course_id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	AvatarURL string    `json:"avatar_url"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}
