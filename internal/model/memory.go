package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Memory 长期记忆模型
type Memory struct {
	ID        string `gorm:"primaryKey;type:varchar(64)" json:"id"`
	PersonaID string `gorm:"index;type:varchar(64);not null" json:"persona_id"`
	UserID    int64  `gorm:"index;not null" json:"user_id"`

	// 记忆内容
	Type     string `gorm:"type:varchar(20);not null" json:"type"` // fact/preference/event/emotion/relationship
	Content  string `gorm:"type:text;not null" json:"content"`
	Keywords string `gorm:"type:varchar(500)" json:"keywords"`

	// 向量嵌入
	Embedding          string     `gorm:"type:longtext" json:"-"`
	EmbeddingUpdatedAt *time.Time `json:"embedding_updated_at,omitempty"`

	// 来源
	Source string `gorm:"type:varchar(20);default:'manual'" json:"source"` // manual/auto

	// 冲突处理
	Status       string `gorm:"type:varchar(20);default:'active'" json:"status"` // active/superseded
	SupersededBy string `gorm:"type:varchar(64)" json:"superseded_by"`

	// 统计
	HitCount  int        `gorm:"default:0" json:"hit_count"`
	LastHitAt *time.Time `json:"last_hit_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `gorm:"default:false" json:"-"`
}

func (m *Memory) TableName() string {
	return "memories"
}

func (m *Memory) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = "mem:" + uuid.New().String()
	return
}

// MemoryType 常量
const (
	MemoryTypeFact         = "fact"
	MemoryTypePreference   = "preference"
	MemoryTypeEvent        = "event"
	MemoryTypeEmotion      = "emotion"
	MemoryTypeRelationship = "relationship"
)

// MemorySource 常量
const (
	MemorySourceManual = "manual"
	MemorySourceAuto   = "auto"
)

// MemoryStatus 常量
const (
	MemoryStatusActive     = "active"
	MemoryStatusSuperseded = "superseded"
)
