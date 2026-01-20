package handler

import (
	"AI_Chat/internal/chat_core"
	"AI_Chat/internal/common"
	"AI_Chat/internal/model"
	"AI_Chat/internal/repository"
	"AI_Chat/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChatHandler struct {
	conversationRepository *repository.ConversationRepository
	personaRepository      *repository.PersonaRepository
}

func NewChatHandler(conversationRepository *repository.ConversationRepository, personaRepository *repository.PersonaRepository) *ChatHandler {
	return &ChatHandler{conversationRepository: conversationRepository, personaRepository: personaRepository}
}
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	var res struct {
		ConversationId string `json:"conversationId"`
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
	conversation := &model.Conversation{
		UserID: userId,
		Title:  req.Title,
	}
	err = h.conversationRepository.CreateConversation(conversation)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	res.ConversationId = conversation.ID
	common.Success(c, res)
}

func (h *ChatHandler) ChatWithPersona(c *gin.Context) {
	var req struct {
		Query          string `json:"query" binding:"required"`
		ConversationId string `json:"conversationId" binding:"required"`
		PersonaId      string `json:"personaId" binding:"required"`
	}
	var res struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.FailedCode)
		return
	}
	persona, err := h.personaRepository.GetPersonaById(req.PersonaId)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	conversation_messages, err := h.conversationRepository.GetMessagesByConversationId(req.ConversationId)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	resp, err := chat_core.Chat(c, req.Query, conversation_messages, persona.SystemPrompt)
	res.Message = resp
	if err != nil {
		utils.Log.Error("聊天失败", zap.Error(err))
		common.Fail(c, common.ChatFailedCode)
		return
	}

	//存放历史消息到mysql，这一步不放在Chat里面，可以根据实际业务灵活操作。
	h.conversationRepository.AddMessageToConversation(
		&model.Message{
			ConversationID: req.ConversationId,
			Role:           "user",
			Content:        req.Query,
			Model:          "deepseek-chat",
			TokenCount:     0,
			CreatedAt:      time.Now(),
		},
	)
	h.conversationRepository.AddMessageToConversation(
		&model.Message{
			ConversationID: req.ConversationId,
			Role:           "assistant",
			Content:        resp,
			Model:          "deepseek-chat",
			TokenCount:     0,
			CreatedAt:      time.Now(),
		},
	)

	common.Success(c, res)
}

func (h *ChatHandler) GetConversations(c *gin.Context) {
	var res struct {
		Conversations []model.Conversation `json:"conversations"`
	}
	userId, err := utils.GetUserIdFromSession(c)
	if err != nil {
		common.Fail(c, common.FailedCode)
		return
	}
	conversations, err := h.conversationRepository.GetConversationsByUserId(userId)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	res.Conversations = conversations
	common.Success(c, res)
}

func (h *ChatHandler) GetConversationMessages(c *gin.Context) {
	var req struct {
		ConversationId string `json:"conversationId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.FailedCode)
		return
	}
	var res struct {
		Messages []model.Message `json:"messages"`
	}
	conversationId := req.ConversationId
	conversationMessages, err := h.conversationRepository.GetMessagesByConversationId(conversationId)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	res.Messages = conversationMessages
	common.Success(c, res)
}
