package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Email     string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Avatar    string         `gorm:"size:255" json:"avatar"`
	Role      string         `gorm:"size:50;default:'user'" json:"role"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Conversations []Conversation `gorm:"foreignKey:UserID" json:"conversations,omitempty"`
	Memories      []Memory      `gorm:"foreignKey:UserID" json:"memories,omitempty"`
	Documents     []Document    `gorm:"foreignKey:UserID" json:"documents,omitempty"`
	Agents        []Agent       `gorm:"foreignKey:UserID" json:"agents,omitempty"`
	Notifications []Notification `gorm:"foreignKey:UserID" json:"notifications,omitempty"`
}

type Conversation struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Title       string         `gorm:"size:255" json:"title"`
	Model       string         `gorm:"size:100" json:"model"`
	IsArchived  bool           `gorm:"default:false" json:"is_archived"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

type Message struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ConversationID uint           `gorm:"not null;index" json:"conversation_id"`
	Role           string         `gorm:"size:20;not null" json:"role"` // user, assistant, system
	Content        string         `gorm:"type:text" json:"content"`
	Tokens         int            `json:"tokens"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
}

type Memory struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Embedding   []float32      `gorm:"type:vector" json:"embedding,omitempty"`
	Importance  float32        `gorm:"default:0.5" json:"importance"`
	AccessCount int            `gorm:"default:0" json:"access_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type Document struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Title       string         `gorm:"size:255" json:"title"`
	FileName    string         `gorm:"size:255;not null" json:"file_name"`
	FilePath    string         `gorm:"size:500;not null" json:"file_path"`
	FileSize    int64          `json:"file_size"`
	FileType    string         `gorm:"size:100" json:"file_type"`
	Status      string         `gorm:"size:50;default:'processing'" json:"status"` // processing, completed, failed
	Summary     string         `gorm:"type:text" json:"summary"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type Agent struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Config      string         `gorm:"type:jsonb" json:"config"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type Notification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Message   string         `gorm:"type:text" json:"message"`
	Type      string         `gorm:"size:50" json:"type"` // email, push, reminder
	IsRead    bool           `gorm:"default:false" json:"is_read"`
	ScheduledAt *time.Time   `json:"scheduled_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
