package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Domain errors
var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrMessageNotFound      = errors.New("message not found")
	ErrInvalidRole          = errors.New("invalid message role")
	ErrEmptyContent         = errors.New("message content cannot be empty")
	ErrConversationClosed   = errors.New("conversation is closed")
	ErrMaxMessagesReached   = errors.New("maximum messages per conversation reached")
)

// Constants
const (
	MaxMessagesPerConversation = 10000
	MaxMessageLength           = 32000 // ~8k tokens
	MaxConversationTitleLength = 100
)

// Conversation represents a chat conversation
type Conversation struct {
	ID            ConversationID     `json:"id" gorm:"primaryKey"`
	UserID        string             `json:"user_id" gorm:"index"`
	Title         string             `json:"title"`
	Model         string             `json:"model"`
	SystemPrompt  string             `json:"system_prompt,omitempty"`
	Messages      []Message          `json:"messages" gorm:"foreignKey:ConversationID"`
	MessageCount  int                `json:"message_count"`
	TokenCount    int                `json:"token_count"`
	Status        ConversationStatus `json:"status"`
	LastMessageAt *time.Time         `json:"last_message_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	DeletedAt     *time.Time         `json:"deleted_at,omitempty" gorm:"index"`
	Version       int                `json:"version" gorm:"column:version;default:1"`
	Metadata      Metadata           `json:"metadata" gorm:"type:jsonb"`
}

// ConversationID is a value object for conversation ID
type ConversationID string

// NewConversationID creates a new conversation ID
func NewConversationID() ConversationID {
	return ConversationID(uuid.New().String())
}

// ParseConversationID parses a string into a ConversationID
func ParseConversationID(id string) (ConversationID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", errors.New("invalid conversation ID")
	}
	return ConversationID(id), nil
}

// String returns the string representation
func (id ConversationID) String() string {
	return string(id)
}

// ConversationStatus represents the conversation status
type ConversationStatus string

const (
	ConversationStatusActive   ConversationStatus = "active"
	ConversationStatusArchived ConversationStatus = "archived"
	ConversationStatusClosed   ConversationStatus = "closed"
)

// IsActive checks if the conversation is active
func (s ConversationStatus) IsActive() bool {
	return s == ConversationStatusActive
}

// Message represents a single message in a conversation
type Message struct {
	ID             MessageID    `json:"id" gorm:"primaryKey"`
	ConversationID string       `json:"conversation_id" gorm:"index"`
	Role           MessageRole  `json:"role"`
	Content        string       `json:"content"`
	TokenCount     int          `json:"token_count"`
	Model          string       `json:"model,omitempty"`
	FinishReason   string       `json:"finish_reason,omitempty"`
	ParentID       *string      `json:"parent_id,omitempty" gorm:"index"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Metadata       Metadata     `json:"metadata" gorm:"type:jsonb"`
	Attachments    []Attachment `json:"attachments,omitempty" gorm:"foreignKey:MessageID"`
}

// MessageID is a value object for message ID
type MessageID string

// NewMessageID creates a new message ID
func NewMessageID() MessageID {
	return MessageID(uuid.New().String())
}

// String returns the string representation
func (id MessageID) String() string {
	return string(id)
}

// MessageRole represents the role of a message
type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleFunction  MessageRole = "function"
)

// IsValid checks if the role is valid
func (r MessageRole) IsValid() bool {
	switch r {
	case MessageRoleSystem, MessageRoleUser, MessageRoleAssistant, MessageRoleFunction:
		return true
	default:
		return false
	}
}

// Attachment represents a file attachment
type Attachment struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	MessageID string    `json:"message_id" gorm:"index"`
	Type      string    `json:"type"` // image, document, etc.
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

// Metadata represents flexible metadata storage
type Metadata map[string]interface{}

// Factory methods

// NewConversation creates a new conversation
func NewConversation(userID, title, model string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:           NewConversationID(),
		UserID:       userID,
		Title:        truncateString(title, MaxConversationTitleLength),
		Model:        model,
		Messages:     make([]Message, 0),
		MessageCount: 0,
		TokenCount:   0,
		Status:       ConversationStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
		Metadata:     make(Metadata),
	}
}

// NewMessage creates a new message
func NewMessage(conversationID string, role MessageRole, content string) (*Message, error) {
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}
	
	if content == "" {
		return nil, ErrEmptyContent
	}
	
	if len(content) > MaxMessageLength {
		content = truncateString(content, MaxMessageLength)
	}
	
	now := time.Now()
	return &Message{
		ID:             NewMessageID(),
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		TokenCount:     estimateTokenCount(content),
		CreatedAt:      now,
		UpdatedAt:      now,
		Metadata:       make(Metadata),
		Attachments:    make([]Attachment, 0),
	}, nil
}

// Business logic methods

// AddMessage adds a message to the conversation
func (c *Conversation) AddMessage(role MessageRole, content string) (*Message, error) {
	if !c.Status.IsActive() {
		return nil, ErrConversationClosed
	}
	
	if c.MessageCount >= MaxMessagesPerConversation {
		return nil, ErrMaxMessagesReached
	}
	
	message, err := NewMessage(c.ID.String(), role, content)
	if err != nil {
		return nil, err
	}
	
	c.Messages = append(c.Messages, *message)
	c.MessageCount++
	c.TokenCount += message.TokenCount
	now := time.Now()
	c.LastMessageAt = &now
	c.UpdatedAt = now
	c.Version++
	
	// Auto-generate title from first user message if not set
	if c.Title == "" && role == MessageRoleUser {
		c.Title = generateTitle(content)
	}
	
	return message, nil
}

// Archive archives the conversation
func (c *Conversation) Archive() {
	c.Status = ConversationStatusArchived
	c.UpdatedAt = time.Now()
	c.Version++
}

// Close closes the conversation
func (c *Conversation) Close() {
	c.Status = ConversationStatusClosed
	now := time.Now()
	c.DeletedAt = &now
	c.UpdatedAt = now
	c.Version++
}

// Reopen reopens a closed conversation
func (c *Conversation) Reopen() error {
	if c.Status != ConversationStatusClosed {
		return errors.New("conversation is not closed")
	}
	
	c.Status = ConversationStatusActive
	c.DeletedAt = nil
	c.UpdatedAt = time.Now()
	c.Version++
	return nil
}

// GetLastMessage returns the last message in the conversation
func (c *Conversation) GetLastMessage() *Message {
	if len(c.Messages) == 0 {
		return nil
	}
	return &c.Messages[len(c.Messages)-1]
}

// GetMessagesByRole returns all messages with a specific role
func (c *Conversation) GetMessagesByRole(role MessageRole) []Message {
	var messages []Message
	for _, msg := range c.Messages {
		if msg.Role == role {
			messages = append(messages, msg)
		}
	}
	return messages
}

// Helper functions

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// generateTitle generates a conversation title from content
func generateTitle(content string) string {
	// Simple implementation - take first line or 50 chars
	maxLen := 50
	if len(content) < maxLen {
		return content
	}
	
	// Try to find a natural break point
	for i := maxLen; i > 0; i-- {
		if content[i] == ' ' || content[i] == '.' || content[i] == '?' || content[i] == '!' {
			return content[:i] + "..."
		}
	}
	
	return content[:maxLen] + "..."
}

// estimateTokenCount estimates the token count for a message
func estimateTokenCount(content string) int {
	// Rough estimate: 1 token ≈ 4 characters
	return len(content) / 4
}

// TableName specifies the table name for GORM
func (Conversation) TableName() string {
	return "conversations"
}

// TableName specifies the table name for GORM
func (Message) TableName() string {
	return "messages"
}

// TableName specifies the table name for GORM
func (Attachment) TableName() string {
	return "attachments"
}