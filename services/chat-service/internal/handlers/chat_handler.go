package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/shopmindai/chat-service/internal/models"
	"github.com/shopmindai/chat-service/internal/repository"
	pb "github.com/shopmindai/chat-service/proto"
)

// ChatHandler handles chat-related HTTP and WebSocket requests
type ChatHandler struct {
	repo      repository.ChatRepository
	redis     *redis.Client
	kafka     *kafka.Writer
	upgrader  websocket.Upgrader
	hub       *Hub
}

// Hub maintains active WebSocket connections
type Hub struct {
	clients    map[string]*Client
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
}

// Client represents a WebSocket connection
type Client struct {
	id       string
	userId   string
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
}

// Message represents a chat message
type Message struct {
	ID               string                 `json:"id"`
	ConversationID   string                 `json:"conversationId"`
	Role             string                 `json:"role"`
	Content          string                 `json:"content"`
	Timestamp        time.Time              `json:"timestamp"`
	IsStreaming      bool                   `json:"isStreaming"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *gorm.DB, redis *redis.Client, kafka *kafka.Writer) *ChatHandler {
	hub := &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	handler := &ChatHandler{
		repo:  repository.NewChatRepository(db),
		redis: redis,
		kafka: kafka,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper CORS check
				return true
			},
		},
		hub: hub,
	}

	// Start hub in background
	go hub.run()

	return handler
}

// WebSocket endpoint for real-time chat
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	userId := c.GetString("userId") // From auth middleware
	
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		id:     uuid.New().String(),
		userId: userId,
		conn:   conn,
		send:   make(chan []byte, 256),
		hub:    h.hub,
	}

	h.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(h)
}

// CreateConversation creates a new chat conversation
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userId := c.GetString("userId")

	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation := &models.Conversation{
		ID:        uuid.New().String(),
		UserID:    userId,
		Title:     req.Title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.repo.CreateConversation(context.Background(), conversation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	// Publish event to Kafka
	event := map[string]interface{}{
		"type":           "conversation.created",
		"conversationId": conversation.ID,
		"userId":         userId,
		"timestamp":      time.Now(),
	}
	h.publishEvent(event)

	c.JSON(http.StatusCreated, conversation)
}

// GetConversations returns user's conversations
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userId := c.GetString("userId")
	
	// Try cache first
	ctx := context.Background()
	cacheKey := "conversations:" + userId
	cached, err := h.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json", []byte(cached))
		return
	}

	conversations, err := h.repo.GetUserConversations(ctx, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	// Cache the result
	data, _ := json.Marshal(conversations)
	h.redis.Set(ctx, cacheKey, data, 5*time.Minute)

	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, conversations)
}

// SendMessage handles sending a message in a conversation
func (h *ChatHandler) SendMessage(c *gin.Context) {
	conversationId := c.Param("conversationId")
	userId := c.GetString("userId")

	var req struct {
		Content string   `json:"content" binding:"required"`
		Files   []string `json:"files,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user message
	userMessage := &models.Message{
		ID:             uuid.New().String(),
		ConversationID: conversationId,
		Role:           "user",
		Content:        req.Content,
		UserID:         userId,
		CreatedAt:      time.Now(),
	}

	if err := h.repo.CreateMessage(context.Background(), userMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	// Broadcast to WebSocket clients
	h.broadcastMessage(userMessage)

	// Simulate AI response (in production, this would call LLM service)
	go h.generateAIResponse(conversationId, userId, req.Content)

	c.JSON(http.StatusCreated, userMessage)
}

// generateAIResponse simulates AI response with streaming
func (h *ChatHandler) generateAIResponse(conversationId, userId, prompt string) {
	// Create assistant message
	assistantMessage := &models.Message{
		ID:             uuid.New().String(),
		ConversationID: conversationId,
		Role:           "assistant",
		Content:        "",
		CreatedAt:      time.Now(),
	}

	// Start streaming
	streamMsg := &Message{
		ID:             assistantMessage.ID,
		ConversationID: conversationId,
		Role:           "assistant",
		Content:        "",
		Timestamp:      time.Now(),
		IsStreaming:    true,
	}

	// Simulate streaming response
	response := "I understand you're asking about: " + prompt + ". Let me help you with that...\n\n"
	
	// Stream character by character
	for i, char := range response {
		streamMsg.Content = response[:i+1]
		h.hub.broadcast <- streamMsg
		time.Sleep(30 * time.Millisecond) // Simulate typing
	}

	// Save complete message
	assistantMessage.Content = response
	h.repo.CreateMessage(context.Background(), assistantMessage)

	// Send completion signal
	streamMsg.IsStreaming = false
	h.hub.broadcast <- streamMsg

	// Publish event
	h.publishEvent(map[string]interface{}{
		"type":           "message.created",
		"conversationId": conversationId,
		"messageId":      assistantMessage.ID,
		"role":           "assistant",
		"timestamp":      time.Now(),
	})
}

// Hub methods
func (hub *Hub) run() {
	for {
		select {
		case client := <-hub.register:
			hub.clients[client.id] = client
			log.Printf("Client %s connected", client.id)

		case client := <-hub.unregister:
			if _, ok := hub.clients[client.id]; ok {
				delete(hub.clients, client.id)
				close(client.send)
				log.Printf("Client %s disconnected", client.id)
			}

		case message := <-hub.broadcast:
			data, _ := json.Marshal(message)
			for _, client := range hub.clients {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(hub.clients, client.id)
				}
			}
		}
	}
}

// Client methods
func (c *Client) readPump(h *ChatHandler) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Process message based on type
		// This is where you'd handle different message types
		log.Printf("Received message from client %s: %v", c.id, msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Helper methods
func (h *ChatHandler) broadcastMessage(msg *models.Message) {
	broadcastMsg := &Message{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		Role:           msg.Role,
		Content:        msg.Content,
		Timestamp:      msg.CreatedAt,
		IsStreaming:    false,
	}
	h.hub.broadcast <- broadcastMsg
}

func (h *ChatHandler) publishEvent(event map[string]interface{}) {
	data, _ := json.Marshal(event)
	err := h.kafka.WriteMessages(context.Background(), kafka.Message{
		Topic: "chat-events",
		Value: data,
	})
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
	}
}