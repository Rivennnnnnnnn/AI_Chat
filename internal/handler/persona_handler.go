package handler

import (
	"AI_Chat/internal/common"
	"AI_Chat/internal/model"
	"AI_Chat/internal/repository"
	"AI_Chat/pkg/utils"

	"github.com/gin-gonic/gin"
)

type PersonaHandler struct {
	personaRepository *repository.PersonaRepository
}

func NewPersonaHandler(personaRepository *repository.PersonaRepository) *PersonaHandler {
	return &PersonaHandler{personaRepository: personaRepository}
}

func (h *PersonaHandler) CreatePersona(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description" binding:"required"`
		SystemPrompt string `json:"systemPrompt" binding:"required"`
		Mode         int    `json:"mode" binding:"required"`
		Avatar       string `json:"avatar" binding:"required"`
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
	persona := &model.Persona{
		UserID:       userId,
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Mode:         req.Mode,
		Avatar:       req.Avatar,
	}
	err = h.personaRepository.CreatePersona(persona)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	common.Success(c, persona)
}

func (h *PersonaHandler) GetPersonas(c *gin.Context) {
	var res struct {
		Personas []model.Persona `json:"personas"`
	}
	userId, err := utils.GetUserIdFromSession(c)
	if err != nil {
		common.Fail(c, common.FailedCode)
		return
	}
	personas, err := h.personaRepository.GetPersonasByUserId(userId)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	res.Personas = personas
	common.Success(c, res)
}
