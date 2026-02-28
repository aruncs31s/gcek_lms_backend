package handler

import (
	"net/http"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	userRepo repository.UserRepository
}

func NewLeaderboardHandler(userRepo repository.UserRepository) *LeaderboardHandler {
	return &LeaderboardHandler{userRepo: userRepo}
}

func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	// Let's get top 50 users
	users, err := h.userRepo.GetLeaderboard(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}

	var res []dto.LeaderboardUserResponse
	for _, u := range users {
		res = append(res, dto.LeaderboardUserResponse{
			UserID:         u.ID.String(),
			FirstName:      u.Profile.FirstName,
			LastName:       u.Profile.LastName,
			AvatarURL:      u.Profile.AvatarURL,
			Points:         u.Profile.Points,
			EnrolledRoutes: len(u.Enrollments), // mapped earlier in user models
		})
	}

	if res == nil {
		res = []dto.LeaderboardUserResponse{}
	}

	c.JSON(http.StatusOK, res)
}
