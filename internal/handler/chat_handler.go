package handler

import (
	"net/http"
	"strings"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/middleware"
	"github.com/aruncs/esdc-lms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chatHub     *service.ChatHub
	chatService service.ChatService
}

func NewChatHandler(hub *service.ChatHub, cs service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatHub:     hub,
		chatService: cs,
	}
}

// ServeWS godoc
// @Summary      WebSocket chat connection
// @Description  Establishes a WebSocket connection for real-time chat. Pass token and user_id as query parameters.
// @Tags         chat
// @Param        token    query  string  true  "JWT token"
// @Param        user_id  query  string  true  "User ID (UUID)"
// @Success      101
// @Failure      401  {object}  map[string]string
// @Router       /ws/chat [get]
func (h *ChatHandler) ServeWS(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	rawUserID := c.Query("user_id")
	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user validation"})
		return
	}

	h.chatHub.ServeWS(c.Writer, c.Request, userID)
}

// CreateConversation godoc
// @Summary      Create a conversation
// @Description  Creates a new direct or group conversation.
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateConversationRequest  true  "Conversation creation payload"
// @Success      201   {object}  dto.ConversationResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/chat/conversations [post]
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userClaims := userClaimsRaw.(middleware.UserClaims)
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.chatService.CreateConversation(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetConversations godoc
// @Summary      Get conversations
// @Description  Returns all conversations for the authenticated user.
// @Tags         chat
// @Produce      json
// @Success      200  {array}   dto.ConversationResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/chat/conversations [get]
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userClaims := userClaimsRaw.(middleware.UserClaims)
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	res, err := h.chatService.GetConversations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetMessages godoc
// @Summary      Get conversation messages
// @Description  Returns all messages in a specific conversation.
// @Tags         chat
// @Produce      json
// @Param        id  path  string  true  "Conversation ID (UUID)"
// @Success      200  {array}   dto.MessageResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/chat/conversations/{id}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userClaimsRaw, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userClaims := userClaimsRaw.(middleware.UserClaims)
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	convIDStr := c.Param("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	res, err := h.chatService.GetMessages(userID, convID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.HasPrefix(err.Error(), "forbidden") {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
