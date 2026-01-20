package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Persona AI角色人格模型
type Persona struct {
	ID           string    `gorm:"primaryKey;column:id;type:varchar(64)" json:"id"`
	UserID       int64     `gorm:"column:user_id;index:idx_user_id" json:"user_id"`
	Name         string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Description  string    `gorm:"column:description;type:text" json:"description"`
	SystemPrompt string    `gorm:"column:system_prompt;type:text" json:"system_prompt"`
	Mode         int       `gorm:"column:mode;type:tinyint;default:1" json:"mode"` // 1:自定义, 2:模拟
	Avatar       string    `gorm:"column:avatar;type:varchar(255)" json:"avatar"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Persona) TableName() string {
	return "personas"
}

func (p *Persona) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = "per:" + uuid.New().String()
	return
}
