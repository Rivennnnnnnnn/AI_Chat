package handler

import (
	"AI_Chat/internal/chat_core"
	"AI_Chat/internal/common"
	"AI_Chat/internal/model"
	"AI_Chat/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	conversationRepository *repository.ConversationRepository
}

func NewChatHandler(conversationRepository *repository.ConversationRepository) *ChatHandler {
	return &ChatHandler{conversationRepository: conversationRepository}
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
	userSession, exists := c.Get("userSession")
	if !exists {
		common.Fail(c, common.SessionExpiredCode)
		return
	}
	userSessionMap, _ := userSession.(map[string]string)
	userId, err := strconv.ParseInt(userSessionMap["id"], 10, 64)
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

func (h *ChatHandler) Chat(c *gin.Context) {
	var req struct {
		Query          string `json:"query" binding:"required"`
		ConversationId string `json:"conversationId" binding:"required"`
		SystemPrompt   string `json:"systemPrompt" binding:"required"`
	}
	var res struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.FailedCode)
		return
	}
	conversation_messages, err := h.conversationRepository.GetMessagesByConversationId(req.ConversationId)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	resp, err := chat_core.Chat(c, req.Query, conversation_messages, req.SystemPrompt)
	res.Message = resp
	if err != nil {
		common.Fail(c, common.ChatFailedCode)
		return
	}
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
