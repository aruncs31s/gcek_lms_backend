package handler

import (
	"net/http"

	"github.com/aruncs/esdc-lms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// GetNotifications godoc
// @Summary      Get notifications
// @Description  Returns all notifications for the authenticated user.
// @Tags         notifications
// @Produce      json
// @Success      200  {array}   dto.NotificationResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/notifications [get]
// GetNotifications returns all notifications for the authenticated user
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	res, err := h.notificationService.GetUserNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetUnreadCount godoc
// @Summary      Get unread notification count
// @Description  Returns the count of unread notifications for the authenticated user.
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  map[string]int64
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/notifications/unread-count [get]
// GetUnreadCount returns the count of unread notifications
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkAsRead godoc
// @Summary      Mark notification as read
// @Description  Marks a specific notification as read for the authenticated user.
// @Tags         notifications
// @Produce      json
// @Param        id  path  string  true  "Notification ID (UUID)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/notifications/{id}/read [put]
// MarkAsRead marks a specific notification as read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, err := getUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	notificationID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.MarkAsRead(notificationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}
