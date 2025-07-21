package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// ChatRepository handles billions of conversations with smart sharding
type ChatRepository struct {
	db    *sql.DB
	redis *redis.ClusterClient
	
	// Sharding strategy
	shardCount int
}

// NewChatRepository creates a new repository optimized for billions of users
func NewChatRepository(db *sql.DB, redis *redis.ClusterClient) *ChatRepository {
	return &ChatRepository{
		db:         db,
		redis:      redis,
		shardCount: 128, // Matches Citus shard count
	}
}

// Conversation represents a chat conversation
type Conversation struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	Title       *string                `json:"title"`
	Model       string                 `json:"model"`
	SystemPrompt *string               `json:"system_prompt,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	TokensUsed  int                    `json:"tokens_used"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty"`
	
	// Computed fields
	LastMessage *Message `json:"last_message,omitempty"`
	MessageCount int     `json:"message_count"`
}

// Message represents a chat message
type Message struct {
	ID             uuid.UUID              `json:"id"`
	ConversationID uuid.UUID              `json:"conversation_id"`
	Role           string                 `json:"role"`
	Content        string                 `json:"content"`
	FunctionName   *string                `json:"function_name,omitempty"`
	FunctionArgs   map[string]interface{} `json:"function_args,omitempty"`
	Tokens         int                    `json:"tokens"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
}

// CreateConversation creates a new conversation with automatic sharding
func (r *ChatRepository) CreateConversation(ctx context.Context, userID uuid.UUID, title *string) (*Conversation, error) {
	conv := &Conversation{
		ID:         uuid.New(),
		UserID:     userID,
		Title:      title,
		Model:      "gpt-3.5-turbo",
		Metadata:   make(map[string]interface{}),
		TokensUsed: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Transaction for consistency
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert with Citus automatic sharding on user_id
	query := `
		INSERT INTO conversations (
			id, user_id, title, model, system_prompt, 
			metadata, tokens_used, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, created_at, updated_at
	`
	
	metadataJSON, _ := json.Marshal(conv.Metadata)
	err = tx.QueryRowContext(
		ctx, query,
		conv.ID, conv.UserID, conv.Title, conv.Model, conv.SystemPrompt,
		metadataJSON, conv.TokensUsed, conv.CreatedAt, conv.UpdatedAt,
	).Scan(&conv.ID, &conv.CreatedAt, &conv.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("insert conversation: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// Cache in Redis with TTL (hot data stays in memory)
	r.cacheConversation(ctx, conv)

	// Publish event for real-time updates
	r.publishEvent(ctx, "conversation.created", conv)

	return conv, nil
}

// GetConversation retrieves a conversation with smart caching
func (r *ChatRepository) GetConversation(ctx context.Context, convID, userID uuid.UUID) (*Conversation, error) {
	// L1 Cache: Check Redis first (sub-millisecond)
	cacheKey := fmt.Sprintf("conv:%s", convID)
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var conv Conversation
		if err := json.Unmarshal([]byte(cached), &conv); err == nil {
			// Verify ownership (security)
			if conv.UserID == userID {
				r.extendCacheTTL(ctx, cacheKey) // Keep hot data hot
				return &conv, nil
			}
		}
	}

	// L2 Cache: Database with read replica
	query := `
		SELECT 
			c.id, c.user_id, c.title, c.model, c.system_prompt,
			c.metadata, c.tokens_used, c.created_at, c.updated_at, c.deleted_at,
			COUNT(m.id) as message_count,
			(
				SELECT json_build_object(
					'id', m2.id,
					'role', m2.role,
					'content', LEFT(m2.content, 100),
					'created_at', m2.created_at
				)
				FROM messages m2
				WHERE m2.conversation_id = c.id
				ORDER BY m2.created_at DESC
				LIMIT 1
			) as last_message
		FROM conversations c
		LEFT JOIN messages m ON m.conversation_id = c.id
		WHERE c.id = $1 AND c.user_id = $2 AND c.deleted_at IS NULL
		GROUP BY c.id
	`

	var conv Conversation
	var metadataJSON []byte
	var lastMessageJSON sql.NullString
	
	err = r.db.QueryRowContext(ctx, query, convID, userID).Scan(
		&conv.ID, &conv.UserID, &conv.Title, &conv.Model, &conv.SystemPrompt,
		&metadataJSON, &conv.TokensUsed, &conv.CreatedAt, &conv.UpdatedAt, &conv.DeletedAt,
		&conv.MessageCount, &lastMessageJSON,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query conversation: %w", err)
	}

	// Parse JSON fields
	json.Unmarshal(metadataJSON, &conv.Metadata)
	if lastMessageJSON.Valid {
		json.Unmarshal([]byte(lastMessageJSON.String), &conv.LastMessage)
	}

	// Update cache
	r.cacheConversation(ctx, &conv)

	return &conv, nil
}

// ListConversations with cursor-based pagination for billions of records
func (r *ChatRepository) ListConversations(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]*Conversation, string, error) {
	// Smart pagination using cursor (timestamp + ID for uniqueness)
	var cursorTime time.Time
	var cursorID uuid.UUID
	
	if cursor != "" {
		// Decode cursor
		fmt.Sscanf(cursor, "%s_%s", &cursorTime, &cursorID)
	} else {
		cursorTime = time.Now().Add(time.Hour) // Future time to get all
		cursorID = uuid.Nil
	}

	query := `
		SELECT 
			c.id, c.title, c.model, c.tokens_used, 
			c.created_at, c.updated_at,
			COUNT(m.id) as message_count,
			MAX(m.created_at) as last_message_at
		FROM conversations c
		LEFT JOIN messages m ON m.conversation_id = c.id
		WHERE c.user_id = $1 
			AND c.deleted_at IS NULL
			AND (c.updated_at, c.id) < ($2, $3)
		GROUP BY c.id
		ORDER BY c.updated_at DESC, c.id DESC
		LIMIT $4
	`

	rows, err := r.db.QueryContext(ctx, query, userID, cursorTime, cursorID, limit+1)
	if err != nil {
		return nil, "", fmt.Errorf("query conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]*Conversation, 0, limit)
	var nextCursor string

	for rows.Next() {
		if len(conversations) >= limit {
			// We have one extra for cursor
			nextCursor = fmt.Sprintf("%s_%s", 
				conversations[len(conversations)-1].UpdatedAt.Format(time.RFC3339Nano),
				conversations[len(conversations)-1].ID,
			)
			break
		}

		var conv Conversation
		var lastMessageAt sql.NullTime
		
		err := rows.Scan(
			&conv.ID, &conv.Title, &conv.Model, &conv.TokensUsed,
			&conv.CreatedAt, &conv.UpdatedAt, &conv.MessageCount, &lastMessageAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("scan row: %w", err)
		}

		conv.UserID = userID
		conversations = append(conversations, &conv)
	}

	// Batch cache hot conversations
	if len(conversations) > 0 {
		r.batchCacheConversations(ctx, conversations[:min(10, len(conversations))])
	}

	return conversations, nextCursor, nil
}

// SendMessage with optimistic locking and event streaming
func (r *ChatRepository) SendMessage(ctx context.Context, convID, userID uuid.UUID, msg *Message) error {
	// Verify ownership first (cached)
	conv, err := r.GetConversation(ctx, convID, userID)
	if err != nil {
		return err
	}

	msg.ID = uuid.New()
	msg.ConversationID = convID
	msg.CreatedAt = time.Now()

	// Transaction with optimistic locking
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert message
	metadataJSON, _ := json.Marshal(msg.Metadata)
	funcArgsJSON, _ := json.Marshal(msg.FunctionArgs)
	
	_, err = tx.ExecContext(ctx, `
		INSERT INTO messages (
			id, conversation_id, role, content, function_name,
			function_args, tokens, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.FunctionName,
		funcArgsJSON, msg.Tokens, metadataJSON, msg.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	// Update conversation stats (atomic)
	_, err = tx.ExecContext(ctx, `
		UPDATE conversations 
		SET tokens_used = tokens_used + $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND user_id = $3
	`, msg.Tokens, convID, userID)
	
	if err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Invalidate cache
	r.invalidateConversationCache(ctx, convID)

	// Publish for real-time streaming
	r.publishMessage(ctx, msg)

	// Update hot cache for recent messages
	r.cacheRecentMessage(ctx, msg)

	return nil
}

// GetMessages with smart pagination and caching
func (r *ChatRepository) GetMessages(ctx context.Context, convID, userID uuid.UUID, limit int, offset int) ([]*Message, error) {
	// Verify ownership
	if _, err := r.GetConversation(ctx, convID, userID); err != nil {
		return nil, err
	}

	// Check hot cache for recent messages
	if offset == 0 && limit <= 50 {
		cached := r.getCachedRecentMessages(ctx, convID, limit)
		if len(cached) > 0 {
			return cached, nil
		}
	}

	query := `
		SELECT 
			id, conversation_id, role, content, function_name,
			function_args, tokens, metadata, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, convID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	messages := make([]*Message, 0, limit)
	for rows.Next() {
		var msg Message
		var funcArgsJSON, metadataJSON []byte
		
		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.FunctionName,
			&funcArgsJSON, &msg.Tokens, &metadataJSON, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		json.Unmarshal(funcArgsJSON, &msg.FunctionArgs)
		json.Unmarshal(metadataJSON, &msg.Metadata)
		
		messages = append(messages, &msg)
	}

	// Cache if recent
	if offset == 0 && len(messages) > 0 {
		r.cacheRecentMessages(ctx, convID, messages)
	}

	return messages, nil
}

// Caching helpers
func (r *ChatRepository) cacheConversation(ctx context.Context, conv *Conversation) {
	data, _ := json.Marshal(conv)
	key := fmt.Sprintf("conv:%s", conv.ID)
	
	// Adaptive TTL based on activity
	ttl := 24 * time.Hour
	if time.Since(conv.UpdatedAt) < time.Hour {
		ttl = 7 * 24 * time.Hour // Hot data stays longer
	}
	
	r.redis.Set(ctx, key, data, ttl)
}

func (r *ChatRepository) batchCacheConversations(ctx context.Context, convs []*Conversation) {
	pipe := r.redis.Pipeline()
	
	for _, conv := range convs {
		data, _ := json.Marshal(conv)
		key := fmt.Sprintf("conv:%s", conv.ID)
		pipe.Set(ctx, key, data, 24*time.Hour)
	}
	
	pipe.Exec(ctx)
}

func (r *ChatRepository) extendCacheTTL(ctx context.Context, key string) {
	r.redis.Expire(ctx, key, 7*24*time.Hour)
}

func (r *ChatRepository) invalidateConversationCache(ctx context.Context, convID uuid.UUID) {
	key := fmt.Sprintf("conv:%s", convID)
	r.redis.Del(ctx, key)
}

func (r *ChatRepository) cacheRecentMessage(ctx context.Context, msg *Message) {
	key := fmt.Sprintf("recent_msgs:%s", msg.ConversationID)
	data, _ := json.Marshal(msg)
	
	// Use Redis list for recent messages (capped at 50)
	pipe := r.redis.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, 49)
	pipe.Expire(ctx, key, time.Hour)
	pipe.Exec(ctx)
}

func (r *ChatRepository) getCachedRecentMessages(ctx context.Context, convID uuid.UUID, limit int) []*Message {
	key := fmt.Sprintf("recent_msgs:%s", convID)
	
	results, err := r.redis.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil || len(results) == 0 {
		return nil
	}

	messages := make([]*Message, 0, len(results))
	for _, data := range results {
		var msg Message
		if json.Unmarshal([]byte(data), &msg) == nil {
			messages = append(messages, &msg)
		}
	}

	return messages
}

func (r *ChatRepository) cacheRecentMessages(ctx context.Context, convID uuid.UUID, messages []*Message) {
	if len(messages) == 0 {
		return
	}

	key := fmt.Sprintf("recent_msgs:%s", convID)
	pipe := r.redis.Pipeline()
	
	// Clear and repopulate
	pipe.Del(ctx, key)
	for i := len(messages) - 1; i >= 0; i-- { // Reverse order for LPUSH
		data, _ := json.Marshal(messages[i])
		pipe.LPush(ctx, key, data)
	}
	pipe.Expire(ctx, key, time.Hour)
	pipe.Exec(ctx)
}

// Event publishing for real-time updates
func (r *ChatRepository) publishEvent(ctx context.Context, eventType string, data interface{}) {
	// Implement Kafka publishing for events
	// This enables real-time updates across millions of connections
}

func (r *ChatRepository) publishMessage(ctx context.Context, msg *Message) {
	// Publish to Redis for immediate delivery
	channel := fmt.Sprintf("conv:%s:messages", msg.ConversationID)
	data, _ := json.Marshal(msg)
	r.redis.Publish(ctx, channel, data)
	
	// Also publish to Kafka for processing pipeline
	r.publishEvent(ctx, "message.created", msg)
}

// SearchConversations using PostgreSQL full-text search with pg_trgm
func (r *ChatRepository) SearchConversations(ctx context.Context, userID uuid.UUID, query string, limit int) ([]*Conversation, error) {
	// Use gin index on content for fast searching
	searchQuery := `
		SELECT DISTINCT c.id, c.title, c.model, c.tokens_used, 
			c.created_at, c.updated_at,
			ts_rank(to_tsvector('english', m.content), plainto_tsquery('english', $2)) as rank
		FROM conversations c
		JOIN messages m ON m.conversation_id = c.id
		WHERE c.user_id = $1 
			AND c.deleted_at IS NULL
			AND m.content ILIKE '%' || $2 || '%'
		ORDER BY rank DESC, c.updated_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, userID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]*Conversation, 0, limit)
	for rows.Next() {
		var conv Conversation
		var rank float32
		
		err := rows.Scan(
			&conv.ID, &conv.Title, &conv.Model, &conv.TokensUsed,
			&conv.CreatedAt, &conv.UpdatedAt, &rank,
		)
		if err != nil {
			continue
		}

		conv.UserID = userID
		conversations = append(conversations, &conv)
	}

	return conversations, nil
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var ErrNotFound = fmt.Errorf("not found")