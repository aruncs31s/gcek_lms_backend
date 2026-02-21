package service

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aruncs/esdc-lms/internal/dto"
	"github.com/aruncs/esdc-lms/internal/model"
	"github.com/aruncs/esdc-lms/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for MVP. In prod, check origin.
	},
}

type ChatHub struct {
	repo       repository.ChatRepository
	clients    map[uuid.UUID]*websocket.Conn // UserID -> Conn
	register   chan *ClientInfo
	unregister chan uuid.UUID
	broadcast  chan BroadcastMessage
	mu         sync.RWMutex
}

type ClientInfo struct {
	UserID uuid.UUID
	Conn   *websocket.Conn
}

type BroadcastMessage struct {
	ConversationID uuid.UUID
	Message        *dto.MessageResponse
}

func NewChatHub(repo repository.ChatRepository) *ChatHub {
	return &ChatHub{
		repo:       repo,
		clients:    make(map[uuid.UUID]*websocket.Conn),
		register:   make(chan *ClientInfo),
		unregister: make(chan uuid.UUID),
		broadcast:  make(chan BroadcastMessage),
	}
}

func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client.Conn
			h.mu.Unlock()
			log.Printf("User %s connected to chat", client.UserID)

		case userID := <-h.unregister:
			h.mu.Lock()
			if conn, ok := h.clients[userID]; ok {
				conn.Close()
				delete(h.clients, userID)
			}
			h.mu.Unlock()
			log.Printf("User %s disconnected from chat", userID)

		case msg := <-h.broadcast:
			// Fetch participants for this conversation to route the message
			// Since we don't have a GetParticipants repo method yet, we fetch all convs for simplicity
			// In production, build a specific query. For MVP, we broadcast to participants.
			// To keep it clean, let's just push it to connected clients who are in this conversation.
			h.mu.RLock()
			for userID, conn := range h.clients {
				if h.repo.IsParticipant(msg.ConversationID, userID) {
					err := conn.WriteJSON(msg.Message)
					if err != nil {
						log.Printf("Error sending message to %s: %v", userID, err)
						conn.Close()
						delete(h.clients, userID)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *ChatHub) ServeWS(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade Error:", err)
		return
	}

	h.register <- &ClientInfo{UserID: userID, Conn: conn}

	// This goroutine reads messages (if clients send via WS instead of HTTP POST)
	go func() {
		defer func() {
			h.unregister <- userID
		}()

		for {
			var msgPayload struct {
				ConversationID string `json:"conversation_id"`
				Content        string `json:"content"`
			}
			err := conn.ReadJSON(&msgPayload)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}

			// Validate and Save
			convID, err := uuid.Parse(msgPayload.ConversationID)
			if err != nil {
				continue
			}

			if !h.repo.IsParticipant(convID, userID) {
				continue // User not in this conversation
			}

			dbMsg := &model.Message{
				ConversationID: convID,
				SenderID:       userID,
				Content:        msgPayload.Content,
			}

			if err := h.repo.CreateMessage(dbMsg); err != nil {
				log.Println("Failed to save message:", err)
				continue
			}

			// Broadcast
			h.broadcast <- BroadcastMessage{
				ConversationID: convID,
				Message: &dto.MessageResponse{
					ID:             dbMsg.ID.String(),
					ConversationID: dbMsg.ConversationID.String(),
					SenderID:       dbMsg.SenderID.String(),
					Content:        dbMsg.Content,
					CreatedAt:      dbMsg.CreatedAt,
				},
			}
		}
	}()
}

// Service Layer interface
type ChatService interface {
	CreateConversation(creatorID uuid.UUID, req *dto.CreateConversationRequest) (*dto.ConversationResponse, error)
	GetConversations(userID uuid.UUID) ([]dto.ConversationResponse, error)
	GetMessages(userID, convID uuid.UUID) ([]dto.MessageResponse, error)
}

type chatService struct {
	repo repository.ChatRepository
}

func NewChatService(repo repository.ChatRepository) ChatService {
	return &chatService{repo: repo}
}

func (s *chatService) CreateConversation(creatorID uuid.UUID, req *dto.CreateConversationRequest) (*dto.ConversationResponse, error) {
	conv := &model.Conversation{
		Type: model.ConversationType(req.Type),
	}

	// Add participants including creator if not present
	parts := make(map[string]bool)
	parts[creatorID.String()] = true

	for _, p := range req.ParticipantIDs {
		parts[p] = true
	}

	for pIDStr := range parts {
		pID, err := uuid.Parse(pIDStr)
		if err == nil {
			conv.Participants = append(conv.Participants, model.ConversationParticipant{
				UserID:   pID,
				JoinedAt: time.Now(),
			})
		}
	}

	if err := s.repo.CreateConversation(conv); err != nil {
		return nil, err
	}

	return &dto.ConversationResponse{
		ID:        conv.ID.String(),
		Type:      string(conv.Type),
		CreatedAt: conv.CreatedAt,
	}, nil
}

func (s *chatService) GetConversations(userID uuid.UUID) ([]dto.ConversationResponse, error) {
	convs, err := s.repo.GetConversationsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var res []dto.ConversationResponse
	for _, c := range convs {
		res = append(res, dto.ConversationResponse{
			ID:        c.ID.String(),
			Type:      string(c.Type),
			CreatedAt: c.CreatedAt,
		})
	}
	if res == nil {
		res = []dto.ConversationResponse{}
	}
	return res, nil
}

func (s *chatService) GetMessages(userID, convID uuid.UUID) ([]dto.MessageResponse, error) {
	if !s.repo.IsParticipant(convID, userID) {
		return nil, errors.New("forbidden: not a participant")
	}

	msgs, err := s.repo.GetMessagesByConversationID(convID)
	if err != nil {
		return nil, err
	}

	var res []dto.MessageResponse
	for _, m := range msgs {
		res = append(res, dto.MessageResponse{
			ID:             m.ID.String(),
			ConversationID: m.ConversationID.String(),
			SenderID:       m.SenderID.String(),
			Content:        m.Content,
			CreatedAt:      m.CreatedAt,
		})
	}
	if res == nil {
		res = []dto.MessageResponse{}
	}
	return res, nil
}
