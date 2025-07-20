package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopgpt/chat-service/internal/ai"
	"github.com/shopgpt/chat-service/internal/ai/mocks"
	"github.com/shopgpt/chat-service/internal/models"
	"github.com/shopgpt/chat-service/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestHub_Run(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer close(hub.register)
	defer close(hub.unregister)

	// Test client registration
	client := &Client{
		ID:     "test-client-1",
		UserID: "user-123",
		send:   make(chan *Message, 256),
	}

	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.clients[client.ID]
	hub.mu.RUnlock()
	assert.True(t, exists, "Client should be registered")

	// Test client unregistration
	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	_, exists = hub.clients[client.ID]
	hub.mu.RUnlock()
	assert.False(t, exists, "Client should be unregistered")
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Register multiple clients
	client1 := &Client{
		ID:     "client-1",
		UserID: "user-123",
		send:   make(chan *Message, 256),
	}
	client2 := &Client{
		ID:     "client-2",
		UserID: "user-123",
		send:   make(chan *Message, 256),
	}
	client3 := &Client{
		ID:     "client-3",
		UserID: "user-456", // Different user
		send:   make(chan *Message, 256),
	}

	hub.register <- client1
	hub.register <- client2
	hub.register <- client3
	time.Sleep(50 * time.Millisecond)

	// Broadcast message for user-123
	msg := &Message{
		ID:      "msg-1",
		Type:    "chat",
		UserID:  "user-123",
		Content: "Test broadcast",
	}

	go func() {
		hub.broadcast <- msg
	}()

	// Client 1 and 2 should receive the message
	select {
	case received := <-client1.send:
		assert.Equal(t, msg.Content, received.Content)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Client 1 didn't receive broadcast")
	}

	select {
	case received := <-client2.send:
		assert.Equal(t, msg.Content, received.Content)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Client 2 didn't receive broadcast")
	}

	// Client 3 should NOT receive the message
	select {
	case <-client3.send:
		t.Fatal("Client 3 shouldn't receive broadcast for different user")
	case <-time.After(100 * time.Millisecond):
		// Expected timeout
	}
}

func TestWebSocketHandler_HandleWebSocket(t *testing.T) {
	tests := []struct {
		name           string
		authUserID     string
		expectedStatus int
	}{
		{
			name:           "successful connection",
			authUserID:     "user-123",
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:           "unauthorized - no user ID",
			authUserID:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := NewHub()
			go hub.Run()

			aiClient := new(mocks.MockAIClient)
			msgRepo := new(mocks.MockMessageRepository)
			logger := zaptest.NewLogger(t)

			handler := NewWebSocketHandler(hub, aiClient, msgRepo, logger)

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				if tt.authUserID != "" {
					ctx = context.WithValue(ctx, "user_id", tt.authUserID)
				}
				handler.HandleWebSocket(w, r.WithContext(ctx))
			}))
			defer server.Close()

			// Convert http:// to ws://
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

			if tt.authUserID != "" {
				// Try to connect
				ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
				if err == nil {
					defer ws.Close()
					
					// Read welcome message
					var msg Message
					err = ws.ReadJSON(&msg)
					assert.NoError(t, err)
					assert.Equal(t, "system", msg.Type)
					assert.Contains(t, msg.Content, "Welcome to ShopGPT")
				}
				assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			} else {
				// Should fail with 401
				_, resp, _ := websocket.DefaultDialer.Dial(wsURL, nil)
				assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestClient_HandleChatMessage(t *testing.T) {
	tests := []struct {
		name          string
		message       *Message
		aiResponse    *ai.ChatResponse
		aiError       error
		expectError   bool
		expectedType  string
		checkProducts bool
	}{
		{
			name: "successful chat response with products",
			message: &Message{
				ID:      "msg-1",
				Type:    "chat",
				Content: "Show me laptops under $1000",
				Store:   "all",
			},
			aiResponse: &ai.ChatResponse{
				Message: "I found 3 great laptops under $1000:",
				Products: []ai.Product{
					{
						ID:    "prod-1",
						Name:  "Dell Laptop",
						Price: 899.99,
						Store: "Amazon",
					},
					{
						ID:    "prod-2",
						Name:  "HP Laptop",
						Price: 799.99,
						Store: "BestBuy",
					},
				},
			},
			expectedType:  "assistant",
			checkProducts: true,
		},
		{
			name: "AI service error",
			message: &Message{
				ID:      "msg-2",
				Type:    "chat",
				Content: "Find products",
			},
			aiError:      assert.AnError,
			expectError:  true,
			expectedType: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			aiClient := new(mocks.MockAIClient)
			msgRepo := new(mocks.MockMessageRepository)
			logger := zaptest.NewLogger(t)

			client := &Client{
				ID:       "test-client",
				UserID:   "user-123",
				send:     make(chan *Message, 256),
				logger:   logger,
				aiClient: aiClient,
				msgRepo:  msgRepo,
			}

			// Set up mocks
			if tt.aiError != nil {
				aiClient.On("GetChatResponse", mock.Anything, tt.message.Content, tt.message.Store).
					Return(nil, tt.aiError)
			} else {
				aiClient.On("GetChatResponse", mock.Anything, tt.message.Content, tt.message.Store).
					Return(tt.aiResponse, nil)
				msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Message")).
					Return(nil)
			}

			// Execute
			go client.handleChatMessage(tt.message)

			// Wait for typing indicator
			select {
			case msg := <-client.send:
				assert.Equal(t, "typing", msg.Type)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("No typing indicator received")
			}

			// Wait for response
			select {
			case msg := <-client.send:
				assert.Equal(t, tt.expectedType, msg.Type)
				
				if tt.expectError {
					assert.NotEmpty(t, msg.Error)
				} else {
					assert.Equal(t, tt.aiResponse.Message, msg.Content)
					if tt.checkProducts {
						assert.Len(t, msg.Products, len(tt.aiResponse.Products))
					}
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatal("No response received")
			}

			aiClient.AssertExpectations(t)
			msgRepo.AssertExpectations(t)
		})
	}
}

func TestClient_HandleSearchMessage(t *testing.T) {
	// Setup
	aiClient := new(mocks.MockAIClient)
	msgRepo := new(mocks.MockMessageRepository)
	logger := zaptest.NewLogger(t)

	client := &Client{
		ID:       "test-client",
		UserID:   "user-123",
		send:     make(chan *Message, 256),
		logger:   logger,
		aiClient: aiClient,
		msgRepo:  msgRepo,
	}

	searchMsg := &Message{
		ID:      "msg-search",
		Type:    "search",
		Content: "gaming laptop",
		Store:   "amazon",
	}

	products := []Product{
		{
			ID:      "prod-1",
			Name:    "ASUS Gaming Laptop",
			Price:   1299.99,
			Store:   "amazon",
			InStock: true,
			Rating:  4.5,
			Reviews: 234,
		},
		{
			ID:      "prod-2",
			Name:    "MSI Gaming Laptop",
			Price:   1499.99,
			Store:   "amazon",
			InStock: false,
			Rating:  4.3,
			Reviews: 156,
		},
	}

	// Convert to AI products
	aiProducts := make([]ai.Product, len(products))
	for i, p := range products {
		aiProducts[i] = ai.Product{
			ID:          p.ID,
			Name:        p.Name,
			Price:       p.Price,
			Store:       p.Store,
			InStock:     p.InStock,
			Rating:      p.Rating,
			ReviewCount: p.Reviews,
		}
	}

	aiClient.On("SearchProducts", mock.Anything, searchMsg.Content, searchMsg.Store).
		Return(aiProducts, nil)
	msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Message")).
		Return(nil)

	// Execute
	go client.handleSearchMessage(searchMsg)

	// Wait for searching indicator
	select {
	case msg := <-client.send:
		assert.Equal(t, "searching", msg.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("No searching indicator received")
	}

	// Wait for results
	select {
	case msg := <-client.send:
		assert.Equal(t, "search_results", msg.Type)
		assert.Len(t, msg.Products, 2)
		assert.Contains(t, msg.Content, "Found 2 products")
		assert.Contains(t, msg.Content, "ASUS Gaming Laptop")
		assert.Contains(t, msg.Content, "✅ In Stock")
		assert.Contains(t, msg.Content, "❌ Out of Stock")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("No search results received")
	}

	aiClient.AssertExpectations(t)
	msgRepo.AssertExpectations(t)
}

func TestClient_SaveMessage(t *testing.T) {
	msgRepo := new(mocks.MockMessageRepository)
	logger := zaptest.NewLogger(t)

	client := &Client{
		ID:      "test-client",
		UserID:  "user-123",
		logger:  logger,
		msgRepo: msgRepo,
	}

	msg := &Message{
		ID:        "msg-123",
		Type:      "chat",
		Content:   "Test message",
		Store:     "amazon",
		Timestamp: time.Now(),
		Products: []Product{
			{
				ID:    "prod-1",
				Name:  "Test Product",
				Price: 99.99,
			},
		},
	}

	msgRepo.On("Create", mock.Anything, mock.MatchedBy(func(dbMsg *models.Message) bool {
		return dbMsg.ID == msg.ID &&
			dbMsg.UserID == client.UserID &&
			dbMsg.Type == msg.Type &&
			dbMsg.Content == msg.Content &&
			dbMsg.Metadata["store"] == msg.Store &&
			dbMsg.Metadata["products"] != nil
	})).Return(nil)

	err := client.saveMessage(msg)
	assert.NoError(t, err)
	msgRepo.AssertExpectations(t)
}

func TestFormatSearchResults(t *testing.T) {
	tests := []struct {
		name     string
		products []Product
		expected []string
	}{
		{
			name:     "no products",
			products: []Product{},
			expected: []string{"No products found"},
		},
		{
			name: "single product",
			products: []Product{
				{
					Name:    "Test Product",
					Price:   99.99,
					Store:   "Amazon",
					InStock: true,
					Rating:  4.5,
					Reviews: 100,
				},
			},
			expected: []string{
				"Found 1 products",
				"Test Product",
				"$99.99 at Amazon",
				"4.5/5.0 (100 reviews)",
				"✅ In Stock",
			},
		},
		{
			name: "more than 5 products",
			products: []Product{
				{Name: "Product 1", Price: 10, Store: "Store1", InStock: true},
				{Name: "Product 2", Price: 20, Store: "Store2", InStock: true},
				{Name: "Product 3", Price: 30, Store: "Store3", InStock: true},
				{Name: "Product 4", Price: 40, Store: "Store4", InStock: true},
				{Name: "Product 5", Price: 50, Store: "Store5", InStock: true},
				{Name: "Product 6", Price: 60, Store: "Store6", InStock: true},
				{Name: "Product 7", Price: 70, Store: "Store7", InStock: true},
			},
			expected: []string{
				"Found 7 products",
				"Product 1",
				"Product 5",
				"...and 2 more products",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSearchResults(tt.products)
			for _, expected := range tt.expected {
				assert.Contains(t, result, expected)
			}
		})
	}
}

// Integration test
func TestWebSocketIntegration(t *testing.T) {
	// Setup
	hub := NewHub()
	go hub.Run()

	aiClient := new(mocks.MockAIClient)
	msgRepo := new(mocks.MockMessageRepository)
	logger := zaptest.NewLogger(t)

	handler := NewWebSocketHandler(hub, aiClient, msgRepo, logger)

	// Set up mocks for integration test
	aiClient.On("GetChatResponse", mock.Anything, "Hello", "all").
		Return(&ai.ChatResponse{
			Message: "Hello! How can I help you find products today?",
		}, nil)

	msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Message")).
		Return(nil)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "user_id", "user-123")
		handler.HandleWebSocket(w, r.WithContext(ctx))
	}))
	defer server.Close()

	// Connect
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Read welcome message
	var welcomeMsg Message
	err = ws.ReadJSON(&welcomeMsg)
	assert.NoError(t, err)
	assert.Equal(t, "system", welcomeMsg.Type)

	// Send chat message
	chatMsg := Message{
		Type:    "chat",
		Content: "Hello",
	}
	err = ws.WriteJSON(chatMsg)
	assert.NoError(t, err)

	// Read acknowledgment
	var ackMsg Message
	err = ws.ReadJSON(&ackMsg)
	assert.NoError(t, err)
	assert.Equal(t, "ack", ackMsg.Type)

	// Read typing indicator
	var typingMsg Message
	err = ws.ReadJSON(&typingMsg)
	assert.NoError(t, err)
	assert.Equal(t, "typing", typingMsg.Type)

	// Read response
	var responseMsg Message
	err = ws.ReadJSON(&responseMsg)
	assert.NoError(t, err)
	assert.Equal(t, "assistant", responseMsg.Type)
	assert.Contains(t, responseMsg.Content, "How can I help you")

	aiClient.AssertExpectations(t)
	msgRepo.AssertExpectations(t)
}

// Concurrent connections test
func TestWebSocketConcurrentConnections(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	aiClient := new(mocks.MockAIClient)
	msgRepo := new(mocks.MockMessageRepository)
	logger := zaptest.NewLogger(t)

	handler := NewWebSocketHandler(hub, aiClient, msgRepo, logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "user_id", "user-"+r.URL.Query().Get("id"))
		handler.HandleWebSocket(w, r.WithContext(ctx))
	}))
	defer server.Close()

	// Connect multiple clients concurrently
	numClients := 10
	var wg sync.WaitGroup
	wg.Add(numClients)

	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer wg.Done()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?id=" + string(rune(clientID))
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			assert.NoError(t, err)
			defer ws.Close()

			// Read welcome message
			var msg Message
			err = ws.ReadJSON(&msg)
			assert.NoError(t, err)
			assert.Equal(t, "system", msg.Type)
		}(i)
	}

	wg.Wait()

	// Check all clients are registered
	hub.mu.RLock()
	assert.Len(t, hub.clients, numClients)
	hub.mu.RUnlock()
}

// Benchmark tests
func BenchmarkHub_Broadcast(b *testing.B) {
	hub := NewHub()
	go hub.Run()

	// Register 100 clients
	for i := 0; i < 100; i++ {
		client := &Client{
			ID:     fmt.Sprintf("client-%d", i),
			UserID: fmt.Sprintf("user-%d", i%10), // 10 different users
			send:   make(chan *Message, 256),
		}
		hub.register <- client
	}
	time.Sleep(100 * time.Millisecond)

	msg := &Message{
		ID:      "bench-msg",
		Type:    "chat",
		UserID:  "user-0",
		Content: "Benchmark message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.broadcast <- msg
	}
}

func BenchmarkClient_SaveMessage(b *testing.B) {
	msgRepo := new(mocks.MockMessageRepository)
	logger := zap.NewNop()

	client := &Client{
		ID:      "bench-client",
		UserID:  "bench-user",
		logger:  logger,
		msgRepo: msgRepo,
	}

	msgRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	msg := &Message{
		ID:        "bench-msg",
		Type:      "chat",
		Content:   "Benchmark message",
		Store:     "amazon",
		Timestamp: time.Now(),
		Products: []Product{
			{ID: "prod-1", Name: "Product 1", Price: 99.99},
			{ID: "prod-2", Name: "Product 2", Price: 199.99},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.saveMessage(msg)
	}
}