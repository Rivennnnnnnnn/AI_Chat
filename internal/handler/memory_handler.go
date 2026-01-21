package handler

import (
	"AI_Chat/internal/common"
	"AI_Chat/internal/memory"
	"AI_Chat/internal/model"
	"AI_Chat/internal/repository"
	"AI_Chat/pkg/utils"

	"github.com/gin-gonic/gin"
)

type MemoryHandler struct {
	memoryRepository  *repository.MemoryRepository
	personaRepository *repository.PersonaRepository
	memoryService     *memory.MemoryService
}

func NewMemoryHandler(memoryRepo *repository.MemoryRepository, personaRepo *repository.PersonaRepository, memoryService *memory.MemoryService) *MemoryHandler {
	return &MemoryHandler{
		memoryRepository:  memoryRepo,
		personaRepository: personaRepo,
		memoryService:     memoryService,
	}
}

// CreateMemory 手动创建记忆
func (h *MemoryHandler) CreateMemory(c *gin.Context) {
	personaId := c.Param("personaId")
	
	var req struct {
		Type    string `json:"type" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.FailedCode)
		return
	}

	userId, err := utils.GetUserIdFromSession(c)
	if err != nil {
		common.Fail(c, common.FailedCode)
		return
	}

	// 验证 persona 归属
	persona, err := h.personaRepository.GetPersonaById(personaId)
	if err != nil || persona.UserID != userId {
		common.Fail(c, common.FailedCode)
		return
	}

	memory := &model.Memory{
		PersonaID: personaId,
		UserID:    userId,
		Type:      req.Type,
		Content:   req.Content,
		Source:    model.MemorySourceManual,
		Status:    model.MemoryStatusActive,
	}
	h.memoryService.PrepareMemoryEmbedding(c.Request.Context(), memory)

	if err := h.memoryRepository.CreateMemory(memory); err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	h.memoryService.UpsertMilvusEmbedding(c.Request.Context(), memory)

	common.Success(c, memory)
}

// GetMemories 获取人格的所有记忆
func (h *MemoryHandler) GetMemories(c *gin.Context) {
	personaId := c.Param("personaId")

	userId, err := utils.GetUserIdFromSession(c)
	if err != nil {
		common.Fail(c, common.FailedCode)
		return
	}

	memories, err := h.memoryRepository.GetActiveMemoriesByPersonaAndUser(personaId, userId)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}

	common.Success(c, gin.H{"memories": memories})
}

// UpdateMemory 更新记忆
func (h *MemoryHandler) UpdateMemory(c *gin.Context) {
	memoryId := c.Param("memoryId")

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.FailedCode)
		return
	}

	userId, err := utils.GetUserIdFromSession(c)
	if err != nil {
		common.Fail(c, common.FailedCode)
		return
	}

	memory, err := h.memoryRepository.GetMemoryById(memoryId)
	if err != nil || memory.UserID != userId {
		common.Fail(c, common.FailedCode)
		return
	}

	memory.Content = req.Content
	h.memoryService.PrepareMemoryEmbedding(c.Request.Context(), memory)
	if err := h.memoryRepository.UpdateMemory(memory); err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	h.memoryService.UpsertMilvusEmbedding(c.Request.Context(), memory)

	common.Success(c, memory)
}

// DeleteMemory 删除记忆
func (h *MemoryHandler) DeleteMemory(c *gin.Context) {
	memoryId := c.Param("memoryId")

	userId, err := utils.GetUserIdFromSession(c)
	if err != nil {
		common.Fail(c, common.FailedCode)
		return
	}

	memory, err := h.memoryRepository.GetMemoryById(memoryId)
	if err != nil || memory.UserID != userId {
		common.Fail(c, common.FailedCode)
		return
	}

	if err := h.memoryRepository.DeleteMemory(memoryId); err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	h.memoryService.DeleteMilvusMemory(c.Request.Context(), memoryId)

	common.Success(c, nil)
}
