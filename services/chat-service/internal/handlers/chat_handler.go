package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/shopmindai/chat-service/internal/models"
	"github.com/shopmindai/chat-service/internal/repository"
	pb "github.com/shopmindai/chat-service/proto"
)

const (
	// WebSocket limits for scale
	maxConnectionsPerUser = 5
	maxMessageSize       = 65536 // 64KB
	writeWait           = 10 * time.Second
	pongWait            = 60 * time.Second
	pingPeriod          = (pongWait * 9) / 10
	maxMessageRate      = 10 // messages per second
)

// ChatHandler handles chat-related HTTP and WebSocket requests
type ChatHandler struct {
	repo      repository.ChatRepository
	redis     *redis.Client
	kafka     *kafka.Writer
	upgrader  websocket.Upgrader
	hub       *Hub
	config    *Config
}

// Config holds handler configuration
type Config struct {
	AllowedOrigins []string
	MaxConnections int64
}

// Hub maintains active WebSocket connections with proper resource management
type Hub struct {
	clients    map[string]*Client
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	
	// Connection tracking
	userConnections map[string]int
	mu             sync.RWMutex
	
	// Metrics
	activeConnections int64
	messagesSent      int64
}

// Client represents a WebSocket connection with rate limiting
type Client struct {
	id       string
	userId   string
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
	limiter  *rate.Limiter
	lastSeen time.Time
	mu       sync.Mutex
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

// NewChatHandler creates a new chat handler with proper configuration
func NewChatHandler(db *gorm.DB, redis *redis.Client, kafka *kafka.Writer, config *Config) *ChatHandler {
	hub := &Hub{
		clients:         make(map[string]*Client),
		broadcast:       make(chan *Message, 1000), // Buffered channel
		register:        make(chan *Client, 100),
		unregister:      make(chan *Client, 100),
		userConnections: make(map[string]int),
	}

	handler := &ChatHandler{
		repo:   repository.NewChatRepository(db),
		redis:  redis,
		kafka:  kafka,
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				// Proper CORS validation
				for _, allowed := range config.AllowedOrigins {
					if origin == allowed {
						return true
					}
				}
				log.Printf("Rejected WebSocket connection from origin: %s", origin)
				return false
			},
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
				// Log upgrade errors
				log.Printf("WebSocket upgrade error: %v", reason)
			},
		},
		hub: hub,
	}

	// Start hub with worker pool
	go hub.runWithWorkers(4) // 4 broadcast workers

	return handler
}

// WebSocket endpoint for real-time chat with proper resource management
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	userId := c.GetString("userId") // From auth middleware
	
	// Check global connection limit
	if atomic.LoadInt64(&h.hub.activeConnections) >= h.config.MaxConnections {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Server at capacity"})
		return
	}
	
	// Check per-user connection limit
	h.hub.mu.RLock()
	userConns := h.hub.userConnections[userId]
	h.hub.mu.RUnlock()
	
	if userConns >= maxConnectionsPerUser {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many connections"})
		return
	}
	
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Configure connection
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	client := &Client{
		id:       uuid.New().String(),
		userId:   userId,
		conn:     conn,
		send:     make(chan []byte, 256),
		hub:      h.hub,
		limiter:  rate.NewLimiter(rate.Limit(maxMessageRate), maxMessageRate*2),
		lastSeen: time.Now(),
	}

	h.hub.register <- client

	// Start goroutines with proper cleanup
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

// Hub methods with worker pool pattern
func (hub *Hub) runWithWorkers(numWorkers int) {
	// Start broadcast workers
	for i := 0; i < numWorkers; i++ {
		go hub.broadcastWorker()
	}
	
	// Cleanup ticker
	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()
	
	for {
		select {
		case client := <-hub.register:
			hub.mu.Lock()
			hub.clients[client.id] = client
			hub.userConnections[client.userId]++
			atomic.AddInt64(&hub.activeConnections, 1)
			hub.mu.Unlock()
			
			log.Printf("Client %s connected (user: %s, total: %d)", 
				client.id, client.userId, atomic.LoadInt64(&hub.activeConnections))

		case client := <-hub.unregister:
			hub.mu.Lock()
			if _, ok := hub.clients[client.id]; ok {
				delete(hub.clients, client.id)
				hub.userConnections[client.userId]--
				if hub.userConnections[client.userId] <= 0 {
					delete(hub.userConnections, client.userId)
				}
				close(client.send)
				atomic.AddInt64(&hub.activeConnections, -1)
			}
			hub.mu.Unlock()
			
			log.Printf("Client %s disconnected (total: %d)", 
				client.id, atomic.LoadInt64(&hub.activeConnections))
		
		case <-cleanupTicker.C:
			hub.cleanupInactiveClients()
		}
	}
}

// Broadcast worker for parallel message processing
func (hub *Hub) broadcastWorker() {
	for message := range hub.broadcast {
		data, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			continue
		}
		
		hub.mu.RLock()
		// Create slice of clients to avoid holding lock during send
		clients := make([]*Client, 0, len(hub.clients))
		for _, client := range hub.clients {
			clients = append(clients, client)
		}
		hub.mu.RUnlock()
		
		// Send to clients without holding lock
		for _, client := range clients {
			select {
			case client.send <- data:
				atomic.AddInt64(&hub.messagesSent, 1)
			default:
				// Client's send channel is full, close it
				hub.unregister <- client
			}
		}
	}
}

// Cleanup inactive clients
func (hub *Hub) cleanupInactiveClients() {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	
	now := time.Now()
	for id, client := range hub.clients {
		if now.Sub(client.lastSeen) > 5*time.Minute {
			delete(hub.clients, id)
			close(client.send)
			log.Printf("Cleaned up inactive client: %s", id)
		}
	}
}

// Client methods with proper resource management
func (c *Client) readPump(h *ChatHandler) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Rate limiting
		if !c.limiter.Allow() {
			log.Printf("Rate limit exceeded for client %s", c.id)
			continue
		}

		// Update last seen
		c.mu.Lock()
		c.lastSeen = time.Now()
		c.mu.Unlock()

		// Handle incoming message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message format from client %s: %v", c.id, err)
			continue
		}

		// Validate message
		if len(msg.Content) > maxMessageSize {
			log.Printf("Message too large from client %s", c.id)
			continue
		}

		// Process message based on type
		ctx := context.WithTimeout(context.Background(), 5*time.Second)
		if err := h.processMessage(ctx, c, &msg); err != nil {
			log.Printf("Error processing message: %v", err)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write message with batching
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message (batching)
			n := len(c.send)
			for i := 0; i < n && i < 10; i++ { // Max 10 messages per batch
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processMessage handles different message types
func (h *ChatHandler) processMessage(ctx context.Context, client *Client, msg *Message) error {
	// Add request ID for tracing
	msg.ID = uuid.New().String()
	msg.Metadata["requestId"] = msg.ID
	msg.Metadata["userId"] = client.userId
	msg.Metadata["timestamp"] = time.Now().Unix()
	
	// Log for debugging
	log.Printf("Processing message %s from user %s", msg.ID, client.userId)
	
	// Here you would implement actual message processing logic
	// For now, just echo back
	h.hub.broadcast <- msg
	
	return nil
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