package dto

type LeaderboardUserResponse struct {
	UserID         string `json:"user_id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	AvatarURL      string `json:"avatar_url"`
	Points         int    `json:"points"`
	EnrolledRoutes int    `json:"enrolled_courses"`
}
