package repository

import (
	"AI_Chat/internal/model"
	"time"

	"gorm.io/gorm"
)

type MemoryRepository struct {
	db *gorm.DB
}

func NewMemoryRepository(db *gorm.DB) *MemoryRepository {
	return &MemoryRepository{db: db}
}

// CreateMemory 创建记忆
func (r *MemoryRepository) CreateMemory(memory *model.Memory) error {
	return r.db.Create(memory).Error
}

// GetMemoryById 根据ID获取记忆
func (r *MemoryRepository) GetMemoryById(id string) (*model.Memory, error) {
	var memory model.Memory
	if err := r.db.Where("id = ? AND is_deleted = false", id).First(&memory).Error; err != nil {
		return nil, err
	}
	return &memory, nil
}

// GetActiveMemoriesByPersonaAndUser 获取某人格某用户的所有活跃记忆
func (r *MemoryRepository) GetActiveMemoriesByPersonaAndUser(personaId string, userId int64) ([]model.Memory, error) {
	var memories []model.Memory
	err := r.db.Where("persona_id = ? AND user_id = ? AND status = ? AND is_deleted = false",
		personaId, userId, model.MemoryStatusActive).
		Order("created_at DESC").
		Find(&memories).Error
	return memories, err
}

// GetMemoriesByPersonaAndUser 获取某人格某用户的所有记忆（包括已被取代的）
func (r *MemoryRepository) GetMemoriesByPersonaAndUser(personaId string, userId int64) ([]model.Memory, error) {
	var memories []model.Memory
	err := r.db.Where("persona_id = ? AND user_id = ? AND is_deleted = false",
		personaId, userId).
		Order("created_at DESC").
		Find(&memories).Error
	return memories, err
}

// GetMemoriesByIDs 根据ID列表获取记忆（过滤 persona/user/active）
func (r *MemoryRepository) GetMemoriesByIDs(personaId string, userId int64, ids []string) ([]model.Memory, error) {
	if len(ids) == 0 {
		return []model.Memory{}, nil
	}
	var memories []model.Memory
	err := r.db.Where("persona_id = ? AND user_id = ? AND id IN ? AND status = ? AND is_deleted = false",
		personaId, userId, ids, model.MemoryStatusActive).
		Find(&memories).Error
	return memories, err
}

// UpdateMemory 更新记忆
func (r *MemoryRepository) UpdateMemory(memory *model.Memory) error {
	return r.db.Save(memory).Error
}

// UpdateMemoryStatus 更新记忆状态
func (r *MemoryRepository) UpdateMemoryStatus(id string, status string, supersededBy string) error {
	return r.db.Model(&model.Memory{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        status,
			"superseded_by": supersededBy,
		}).Error
}

// UpdateMemoryEmbedding 更新记忆向量
func (r *MemoryRepository) UpdateMemoryEmbedding(id string, embedding string) error {
	return r.db.Model(&model.Memory{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"embedding":            embedding,
			"embedding_updated_at": time.Now(),
		}).Error
}

// DeleteMemory 软删除记忆
func (r *MemoryRepository) DeleteMemory(id string) error {
	return r.db.Model(&model.Memory{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

// IncrementHitCount 增加命中次数
func (r *MemoryRepository) IncrementHitCount(id string) error {
	return r.db.Model(&model.Memory{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"hit_count":   gorm.Expr("hit_count + 1"),
			"last_hit_at": gorm.Expr("NOW()"),
		}).Error
}

// BatchCreateMemories 批量创建记忆
func (r *MemoryRepository) BatchCreateMemories(memories []model.Memory) error {
	if len(memories) == 0 {
		return nil
	}
	return r.db.Create(&memories).Error
}
