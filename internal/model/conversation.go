package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Conversation struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID    int64     `gorm:"not null;index" json:"userId"`
	PersonaID string    `gorm:"index" json:"personaId"`
	Title     string    `gorm:"size:255" json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsDeleted bool      `gorm:"default:false" json:"-"`
}

func (c *Conversation) TableName() string {
	return "conversations"
}
func (c *Conversation) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = "con:" + uuid.New().String()
	return
}

type Message struct {
	OrderID        int64     `gorm:"autoIncrement;not null;index" json:"orderId"`
	ID             string    `gorm:"primaryKey" json:"id"`
	ConversationID string    `gorm:"not null;index" json:"conversationId"`
	Role           string    `gorm:"type:varchar(20);not null" json:"role"` // user, assistant, system
	Content        string    `gorm:"type:text;not null" json:"content"`
	Model          string    `gorm:"size:50" json:"model"`
	TokenCount     int       `json:"tokenCount"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (m *Message) TableName() string {
	return "messages"
}
func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = "msg:" + uuid.New().String()
	return
}
