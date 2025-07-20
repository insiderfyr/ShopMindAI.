package domain

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Domain errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrInvalidUsername   = errors.New("invalid username format")
	ErrInvalidPassword   = errors.New("password must be at least 8 characters")
	ErrInvalidUserID     = errors.New("invalid user ID")
	ErrUserDeactivated   = errors.New("user is deactivated")
)

// User represents the core user entity
type User struct {
	ID              UserID      `json:"id" gorm:"primaryKey"`
	Email           Email       `json:"email" gorm:"uniqueIndex"`
	Username        Username    `json:"username" gorm:"uniqueIndex"`
	DisplayName     string      `json:"display_name"`
	Avatar          string      `json:"avatar"`
	Bio             string      `json:"bio"`
	Preferences     Preferences `json:"preferences" gorm:"type:jsonb"`
	Status          UserStatus  `json:"status"`
	EmailVerified   bool        `json:"email_verified"`
	LastLoginAt     *time.Time  `json:"last_login_at"`
	PasswordHash    string      `json:"-" gorm:"column:password_hash"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	DeletedAt       *time.Time  `json:"deleted_at,omitempty" gorm:"index"`
	Version         int         `json:"version" gorm:"column:version;default:1"`
}

// UserID is a value object for user ID
type UserID string

// NewUserID creates a new user ID
func NewUserID() UserID {
	return UserID(uuid.New().String())
}

// ParseUserID parses a string into a UserID
func ParseUserID(id string) (UserID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidUserID
	}
	return UserID(id), nil
}

// String returns the string representation
func (id UserID) String() string {
	return string(id)
}

// Email is a value object for email addresses
type Email string

// NewEmail creates a new email with validation
func NewEmail(email string) (Email, error) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return "", ErrInvalidEmail
	}
	return Email(email), nil
}

// String returns the string representation
func (e Email) String() string {
	return string(e)
}

// Username is a value object for usernames
type Username string

// NewUsername creates a new username with validation
func NewUsername(username string) (Username, error) {
	if len(username) < 3 || len(username) > 30 {
		return "", ErrInvalidUsername
	}
	
	// Only alphanumeric, underscore, and hyphen allowed
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	if !usernameRegex.MatchString(username) {
		return "", ErrInvalidUsername
	}
	
	return Username(username), nil
}

// String returns the string representation
func (u Username) String() string {
	return string(u)
}

// UserStatus represents the user's account status
type UserStatus string

const (
	UserStatusActive      UserStatus = "active"
	UserStatusInactive    UserStatus = "inactive"
	UserStatusSuspended   UserStatus = "suspended"
	UserStatusDeactivated UserStatus = "deactivated"
)

// IsActive checks if the user status allows activity
func (s UserStatus) IsActive() bool {
	return s == UserStatusActive
}

// Preferences represents user preferences
type Preferences struct {
	Theme              string                 `json:"theme"`
	Language           string                 `json:"language"`
	Timezone           string                 `json:"timezone"`
	EmailNotifications bool                   `json:"email_notifications"`
	PushNotifications  bool                   `json:"push_notifications"`
	Privacy            PrivacySettings        `json:"privacy"`
	ChatSettings       ChatSettings           `json:"chat_settings"`
	Custom             map[string]interface{} `json:"custom"`
}

// PrivacySettings represents privacy preferences
type PrivacySettings struct {
	ProfileVisibility   string `json:"profile_visibility"` // public, friends, private
	ShowOnlineStatus    bool   `json:"show_online_status"`
	AllowDirectMessages bool   `json:"allow_direct_messages"`
}

// ChatSettings represents chat-specific preferences
type ChatSettings struct {
	SaveHistory          bool   `json:"save_history"`
	StreamingResponses   bool   `json:"streaming_responses"`
	DefaultModel         string `json:"default_model"`
	Temperature          float32 `json:"temperature"`
	MaxTokens            int    `json:"max_tokens"`
	SystemPrompt         string `json:"system_prompt"`
}

// DefaultPreferences returns default user preferences
func DefaultPreferences() Preferences {
	return Preferences{
		Theme:              "system",
		Language:           "en",
		Timezone:           "UTC",
		EmailNotifications: true,
		PushNotifications:  true,
		Privacy: PrivacySettings{
			ProfileVisibility:   "public",
			ShowOnlineStatus:    true,
			AllowDirectMessages: true,
		},
		ChatSettings: ChatSettings{
			SaveHistory:        true,
			StreamingResponses: true,
			DefaultModel:       "gpt-4",
			Temperature:        0.7,
			MaxTokens:          2048,
			SystemPrompt:       "",
		},
		Custom: make(map[string]interface{}),
	}
}

// Scan implements sql.Scanner for Preferences
func (p *Preferences) Scan(value interface{}) error {
	if value == nil {
		*p = DefaultPreferences()
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-byte value into Preferences")
	}
	
	return json.Unmarshal(bytes, p)
}

// Value implements driver.Valuer for Preferences
func (p Preferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Factory methods

// NewUser creates a new user with validation
func NewUser(email, username, displayName string) (*User, error) {
	emailObj, err := NewEmail(email)
	if err != nil {
		return nil, err
	}
	
	usernameObj, err := NewUsername(username)
	if err != nil {
		return nil, err
	}
	
	if displayName == "" {
		displayName = username
	}
	
	now := time.Now()
	return &User{
		ID:            NewUserID(),
		Email:         emailObj,
		Username:      usernameObj,
		DisplayName:   displayName,
		Preferences:   DefaultPreferences(),
		Status:        UserStatusActive,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}, nil
}

// Business logic methods

// CanLogin checks if the user can login
func (u *User) CanLogin() error {
	if !u.Status.IsActive() {
		return ErrUserDeactivated
	}
	return nil
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(displayName, bio, avatar string) {
	if displayName != "" {
		u.DisplayName = displayName
	}
	u.Bio = bio
	u.Avatar = avatar
	u.UpdatedAt = time.Now()
	u.Version++
}

// UpdatePreferences updates user preferences
func (u *User) UpdatePreferences(prefs Preferences) {
	u.Preferences = prefs
	u.UpdatedAt = time.Now()
	u.Version++
}

// RecordLogin records a successful login
func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.Status = UserStatusDeactivated
	now := time.Now()
	u.DeletedAt = &now
	u.UpdatedAt = now
	u.Version++
}

// Reactivate reactivates a deactivated user account
func (u *User) Reactivate() error {
	if u.Status != UserStatusDeactivated {
		return errors.New("user is not deactivated")
	}
	
	u.Status = UserStatusActive
	u.DeletedAt = nil
	u.UpdatedAt = time.Now()
	u.Version++
	return nil
}

// VerifyEmail marks the email as verified
func (u *User) VerifyEmail() {
	u.EmailVerified = true
	u.UpdatedAt = time.Now()
	u.Version++
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}