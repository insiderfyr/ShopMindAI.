package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopgpt/chat-service/internal/ai"
	"github.com/shopgpt/chat-service/internal/models"
	"github.com/shopgpt/chat-service/internal/repository"
	"github.com/shopgpt/chat-service/pkg/logger"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, check origin properly
		return true
	},
}

type Hub struct {
	clients    map[string]*Client
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Client struct {
	ID       string
	UserID   string
	conn     *websocket.Conn
	send     chan *Message
	hub      *Hub
	logger   *zap.Logger
	aiClient ai.Client
	msgRepo  repository.MessageRepository
}

type Message struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	UserID    string    `json:"userId"`
	Content   string    `json:"content"`
	Store     string    `json:"store,omitempty"`
	Products  []Product `json:"products,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Store       string  `json:"store"`
	URL         string  `json:"url"`
	Image       string  `json:"image"`
	Description string  `json:"description"`
	InStock     bool    `json:"inStock"`
	Rating      float64 `json:"rating,omitempty"`
	Reviews     int     `json:"reviews,omitempty"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			
			logger.Info("Client connected", 
				zap.String("client_id", client.ID),
				zap.String("user_id", client.UserID))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
				h.mu.Unlock()
				
				logger.Info("Client disconnected",
					zap.String("client_id", client.ID),
					zap.String("user_id", client.UserID))
			} else {
				h.mu.Unlock()
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				if client.UserID == message.UserID {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client.ID)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

type WebSocketHandler struct {
	hub      *Hub
	logger   *zap.Logger
	aiClient ai.Client
	msgRepo  repository.MessageRepository
}

func NewWebSocketHandler(hub *Hub, aiClient ai.Client, msgRepo repository.MessageRepository, logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:      hub,
		logger:   logger,
		aiClient: aiClient,
		msgRepo:  msgRepo,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Create client
	client := &Client{
		ID:       generateClientID(),
		UserID:   userID,
		conn:     conn,
		send:     make(chan *Message, 256),
		hub:      h.hub,
		logger:   h.logger,
		aiClient: h.aiClient,
		msgRepo:  h.msgRepo,
	}

	// Register client
	client.hub.register <- client

	// Send welcome message
	welcomeMsg := &Message{
		ID:        generateMessageID(),
		Type:      "system",
		Content:   "Welcome to ShopGPT! I'm here to help you find the best products across multiple stores. What are you looking for today?",
		Timestamp: time.Now(),
	}
	client.send <- welcomeMsg

	// Start goroutines
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		// Set message metadata
		msg.ID = generateMessageID()
		msg.UserID = c.UserID
		msg.Timestamp = time.Now()

		// Save user message
		if err := c.saveMessage(&msg); err != nil {
			c.logger.Error("Failed to save message", zap.Error(err))
		}

		// Send acknowledgment
		ackMsg := &Message{
			ID:        generateMessageID(),
			Type:      "ack",
			Content:   msg.ID,
			Timestamp: time.Now(),
		}
		c.send <- ackMsg

		// Process message based on type
		switch msg.Type {
		case "chat":
			go c.handleChatMessage(&msg)
		case "search":
			go c.handleSearchMessage(&msg)
		case "typing":
			// Broadcast typing indicator
			c.hub.broadcast <- &msg
		default:
			c.logger.Warn("Unknown message type", zap.String("type", msg.Type))
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				c.logger.Error("Failed to write message", zap.Error(err))
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

func (c *Client) handleChatMessage(msg *Message) {
	ctx := context.Background()

	// Send typing indicator
	typingMsg := &Message{
		ID:        generateMessageID(),
		Type:      "typing",
		Content:   "assistant",
		Timestamp: time.Now(),
	}
	c.send <- typingMsg

	// Get AI response
	response, err := c.aiClient.GetChatResponse(ctx, msg.Content, msg.Store)
	if err != nil {
		c.logger.Error("Failed to get AI response", zap.Error(err))
		errorMsg := &Message{
			ID:        generateMessageID(),
			Type:      "error",
			Error:     "Sorry, I'm having trouble processing your request. Please try again.",
			Timestamp: time.Now(),
		}
		c.send <- errorMsg
		return
	}

	// Parse products if any
	products := extractProductsFromResponse(response)

	// Create response message
	responseMsg := &Message{
		ID:        generateMessageID(),
		Type:      "assistant",
		Content:   response.Message,
		Products:  products,
		Timestamp: time.Now(),
	}

	// Save assistant message
	if err := c.saveMessage(responseMsg); err != nil {
		c.logger.Error("Failed to save assistant message", zap.Error(err))
	}

	// Send response
	c.send <- responseMsg
}

func (c *Client) handleSearchMessage(msg *Message) {
	ctx := context.Background()

	// Send searching indicator
	searchingMsg := &Message{
		ID:        generateMessageID(),
		Type:      "searching",
		Content:   "Searching for products...",
		Timestamp: time.Now(),
	}
	c.send <- searchingMsg

	// Perform search
	products, err := c.aiClient.SearchProducts(ctx, msg.Content, msg.Store)
	if err != nil {
		c.logger.Error("Failed to search products", zap.Error(err))
		errorMsg := &Message{
			ID:        generateMessageID(),
			Type:      "error",
			Error:     "Failed to search products. Please try again.",
			Timestamp: time.Now(),
		}
		c.send <- errorMsg
		return
	}

	// Create response
	responseMsg := &Message{
		ID:        generateMessageID(),
		Type:      "search_results",
		Content:   formatSearchResults(products),
		Products:  products,
		Timestamp: time.Now(),
	}

	// Save search results
	if err := c.saveMessage(responseMsg); err != nil {
		c.logger.Error("Failed to save search results", zap.Error(err))
	}

	// Send results
	c.send <- responseMsg
}

func (c *Client) saveMessage(msg *Message) error {
	ctx := context.Background()
	
	dbMsg := &models.Message{
		ID:        msg.ID,
		UserID:    c.UserID,
		Type:      msg.Type,
		Content:   msg.Content,
		Metadata:  map[string]interface{}{},
		CreatedAt: msg.Timestamp,
	}

	if msg.Store != "" {
		dbMsg.Metadata["store"] = msg.Store
	}

	if len(msg.Products) > 0 {
		productsJSON, _ := json.Marshal(msg.Products)
		dbMsg.Metadata["products"] = string(productsJSON)
	}

	return c.msgRepo.Create(ctx, dbMsg)
}

// Helper functions
func generateClientID() string {
	return "client-" + generateID()
}

func generateMessageID() string {
	return "msg-" + generateID()
}

func generateID() string {
	// In production, use a proper UUID generator
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func extractProductsFromResponse(response *ai.ChatResponse) []Product {
	// Extract products from AI response
	// This would parse structured data from the AI response
	products := []Product{}
	
	if response.Products != nil {
		for _, p := range response.Products {
			products = append(products, Product{
				ID:          p.ID,
				Name:        p.Name,
				Price:       p.Price,
				Store:       p.Store,
				URL:         p.URL,
				Image:       p.ImageURL,
				Description: p.Description,
				InStock:     p.InStock,
				Rating:      p.Rating,
				Reviews:     p.ReviewCount,
			})
		}
	}
	
	return products
}

func formatSearchResults(products []Product) string {
	if len(products) == 0 {
		return "No products found matching your search."
	}

	result := fmt.Sprintf("Found %d products:\n\n", len(products))
	
	for i, p := range products {
		if i >= 5 { // Limit to 5 products in text
			result += fmt.Sprintf("\n...and %d more products", len(products)-5)
			break
		}
		
		result += fmt.Sprintf("%d. **%s**\n", i+1, p.Name)
		result += fmt.Sprintf("   💰 $%.2f at %s\n", p.Price, p.Store)
		if p.Rating > 0 {
			result += fmt.Sprintf("   ⭐ %.1f/5.0 (%d reviews)\n", p.Rating, p.Reviews)
		}
		if p.InStock {
			result += "   ✅ In Stock\n"
		} else {
			result += "   ❌ Out of Stock\n"
		}
		result += "\n"
	}
	
	return result
}

// Constants
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)