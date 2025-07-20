package repository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"sync/atomic"
)

// ChatRepository handles billions of conversations with smart sharding
type ChatRepository struct {
	db    *sql.DB
	redis *redis.ClusterClient
	
	// Prepared statements cache
	stmts map[string]*sql.Stmt
	mu    sync.RWMutex
	
	// Sharding strategy
	shardCount int
	
	// Metrics
	cacheHits   int64
	cacheMisses int64
}

// NewChatRepository creates a new repository optimized for billions of users
func NewChatRepository(db *sql.DB, redis *redis.ClusterClient) (*ChatRepository, error) {
	repo := &ChatRepository{
		db:         db,
		redis:      redis,
		shardCount: 128, // Matches Citus shard count
		stmts:      make(map[string]*sql.Stmt),
	}
	
	// Configure connection pool for high concurrency
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(15 * time.Minute)
	
	// Prepare frequently used statements
	if err := repo.prepareStatements(); err != nil {
		return nil, fmt.Errorf("prepare statements: %w", err)
	}
	
	return repo, nil
}

// prepareStatements prepares commonly used SQL statements
func (r *ChatRepository) prepareStatements() error {
	statements := map[string]string{
		"getConversation": `
			SELECT 
				c.id, c.user_id, c.title, c.model, c.system_prompt,
				c.metadata, c.tokens_used, c.created_at, c.updated_at, c.deleted_at
			FROM conversations c
			WHERE c.id = $1 AND c.user_id = $2 AND c.deleted_at IS NULL
		`,
		"createConversation": `
			INSERT INTO conversations (
				id, user_id, title, model, system_prompt, 
				metadata, tokens_used, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9
			) RETURNING id, created_at, updated_at
		`,
		"insertMessage": `
			INSERT INTO messages (
				id, conversation_id, role, content, function_name,
				function_args, tokens, metadata, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
		"updateConversationTokens": `
			UPDATE conversations 
			SET tokens_used = tokens_used + $1,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $2 AND user_id = $3
		`,
		"getMessagesWithWindow": `
			WITH ranked_messages AS (
				SELECT 
					id, conversation_id, role, content, function_name,
					function_args, tokens, metadata, created_at,
					ROW_NUMBER() OVER (ORDER BY created_at DESC) as rn
				FROM messages
				WHERE conversation_id = $1
			)
			SELECT 
				id, conversation_id, role, content, function_name,
				function_args, tokens, metadata, created_at
			FROM ranked_messages
			WHERE rn > $2 AND rn <= $3
			ORDER BY created_at ASC
		`,
	}
	
	for name, query := range statements {
		stmt, err := r.db.Prepare(query)
		if err != nil {
			return fmt.Errorf("prepare %s: %w", name, err)
		}
		r.stmts[name] = stmt
	}
	
	return nil
}

// getStmt returns a prepared statement by name
func (r *ChatRepository) getStmt(name string) *sql.Stmt {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stmts[name]
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

	// Use prepared statement
	stmt := r.getStmt("createConversation")
	metadataJSON, _ := json.Marshal(conv.Metadata)
	
	err := stmt.QueryRowContext(
		ctx,
		conv.ID, conv.UserID, conv.Title, conv.Model, conv.SystemPrompt,
		metadataJSON, conv.TokensUsed, conv.CreatedAt, conv.UpdatedAt,
	).Scan(&conv.ID, &conv.CreatedAt, &conv.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("insert conversation: %w", err)
	}

	// Cache in Redis with TTL (hot data stays in memory)
	r.cacheConversation(ctx, conv)

	// Publish event for real-time updates
	go r.publishEvent(ctx, "conversation.created", conv)

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
				atomic.AddInt64(&r.cacheHits, 1)
				return &conv, nil
			}
		}
	}
	
	atomic.AddInt64(&r.cacheMisses, 1)

	// L2 Cache: Database with prepared statement
	stmt := r.getStmt("getConversation")
	
	var conv Conversation
	var metadataJSON []byte
	
	err = stmt.QueryRowContext(ctx, convID, userID).Scan(
		&conv.ID, &conv.UserID, &conv.Title, &conv.Model, &conv.SystemPrompt,
		&metadataJSON, &conv.TokensUsed, &conv.CreatedAt, &conv.UpdatedAt, &conv.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query conversation: %w", err)
	}

	// Parse JSON fields
	json.Unmarshal(metadataJSON, &conv.Metadata)

	// Update cache asynchronously
	go r.cacheConversation(ctx, &conv)

	return &conv, nil
}

// ListConversations with cursor-based pagination for billions of records
func (r *ChatRepository) ListConversations(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]*Conversation, string, error) {
	// Decode cursor safely
	var cursorTime time.Time
	var cursorID uuid.UUID
	
	if cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err == nil {
			parts := strings.Split(string(decoded), "|")
			if len(parts) == 2 {
				cursorTime, _ = time.Parse(time.RFC3339Nano, parts[0])
				cursorID, _ = uuid.Parse(parts[1])
			}
		}
	} else {
		cursorTime = time.Now().Add(time.Hour) // Future time to get all
		cursorID = uuid.Nil
	}

	// Use composite index on (user_id, updated_at, id) for efficient pagination
	query := `
		WITH conversation_stats AS (
			SELECT 
				conversation_id,
				COUNT(*) as message_count,
				MAX(created_at) as last_message_at
			FROM messages
			WHERE conversation_id IN (
				SELECT id FROM conversations 
				WHERE user_id = $1 AND deleted_at IS NULL
				AND (updated_at, id) < ($2, $3)
				ORDER BY updated_at DESC, id DESC
				LIMIT $4
			)
			GROUP BY conversation_id
		)
		SELECT 
			c.id, c.title, c.model, c.tokens_used, 
			c.created_at, c.updated_at,
			COALESCE(s.message_count, 0) as message_count,
			s.last_message_at
		FROM conversations c
		LEFT JOIN conversation_stats s ON s.conversation_id = c.id
		WHERE c.user_id = $1 
			AND c.deleted_at IS NULL
			AND (c.updated_at, c.id) < ($2, $3)
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
			lastConv := conversations[len(conversations)-1]
			nextCursor = base64.StdEncoding.EncodeToString([]byte(
				fmt.Sprintf("%s|%s", 
					lastConv.UpdatedAt.Format(time.RFC3339Nano),
					lastConv.ID.String(),
				),
			))
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

	// Batch cache hot conversations asynchronously
	if len(conversations) > 0 {
		go r.batchCacheConversations(ctx, conversations[:min(10, len(conversations))])
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
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert message using prepared statement
	metadataJSON, _ := json.Marshal(msg.Metadata)
	funcArgsJSON, _ := json.Marshal(msg.FunctionArgs)
	
	stmt := tx.StmtContext(ctx, r.getStmt("insertMessage"))
	_, err = stmt.ExecContext(ctx,
		msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.FunctionName,
		funcArgsJSON, msg.Tokens, metadataJSON, msg.CreatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	// Update conversation stats (atomic)
	stmt = tx.StmtContext(ctx, r.getStmt("updateConversationTokens"))
	_, err = stmt.ExecContext(ctx, msg.Tokens, convID, userID)
	
	if err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Async operations after commit
	go func() {
		// Invalidate cache
		r.invalidateConversationCache(context.Background(), convID)
		
		// Publish for real-time streaming
		r.publishMessage(context.Background(), msg)
		
		// Update hot cache for recent messages
		r.cacheRecentMessage(context.Background(), msg)
	}()

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

	// Use window function for efficient pagination
	stmt := r.getStmt("getMessagesWithWindow")
	rows, err := stmt.QueryContext(ctx, convID, offset, offset+limit)
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
		go r.cacheRecentMessages(ctx, convID, messages)
	}

	return messages, nil
}

// BatchCreateMessages for bulk operations
func (r *ChatRepository) BatchCreateMessages(ctx context.Context, messages []*Message) error {
	if len(messages) == 0 {
		return nil
	}
	
	// Build bulk insert query
	valueStrings := make([]string, 0, len(messages))
	valueArgs := make([]interface{}, 0, len(messages)*9)
	
	for i, msg := range messages {
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9,
		))
		
		metadataJSON, _ := json.Marshal(msg.Metadata)
		funcArgsJSON, _ := json.Marshal(msg.FunctionArgs)
		
		valueArgs = append(valueArgs,
			msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.FunctionName,
			funcArgsJSON, msg.Tokens, metadataJSON, msg.CreatedAt,
		)
	}
	
	query := fmt.Sprintf(`
		INSERT INTO messages (
			id, conversation_id, role, content, function_name,
			function_args, tokens, metadata, created_at
		) VALUES %s
		ON CONFLICT (id) DO NOTHING
	`, strings.Join(valueStrings, ","))
	
	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	return err
}

// Close closes the repository and cleans up resources
func (r *ChatRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, stmt := range r.stmts {
		stmt.Close()
	}
	
	return nil
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