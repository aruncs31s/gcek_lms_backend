package dto

import "time"

type AchievementResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	Points      int       `json:"points"`
	EarnedAt    time.Time `json:"earned_at"`
}

type EnrolmentResponse struct {
	CourseID           string    `json:"course_id"`
	CourseTitle        string    `json:"course_title"`
	CourseThumbnailURL string    `json:"course_thumbnail_url"`
	Status             string    `json:"status"`
	ProgressPercentage float64   `json:"progress_percentage"`
	EnrolledAt         time.Time `json:"enrolled_at"`
}

type UserProfileEnrolmentsResponse struct {
	User         UserResponse          `json:"user"`
	Points       int                   `json:"points"`
	Achievements []AchievementResponse `json:"achievements"`
	Enrolments   []EnrolmentResponse   `json:"enrolments"`
	TotalCount   int64                 `json:"total_enrolments"` // if supporting pagination on enrolments
}
