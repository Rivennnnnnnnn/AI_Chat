package handler

import (
	"AI_Chat/internal/chat_core"
	"AI_Chat/internal/chat_core/llm_tools"
	"AI_Chat/internal/common"
	"AI_Chat/internal/memory"
	"AI_Chat/internal/model"
	"AI_Chat/internal/repository"
	"AI_Chat/pkg/utils"
	"context"
	"errors"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ChatHandler struct {
	conversationRepository *repository.ConversationRepository
	personaRepository      *repository.PersonaRepository
	memoryService          *memory.MemoryService
}

func NewChatHandler(conversationRepository *repository.ConversationRepository, personaRepository *repository.PersonaRepository, memoryService *memory.MemoryService) *ChatHandler {
	return &ChatHandler{
		conversationRepository: conversationRepository,
		personaRepository:      personaRepository,
		memoryService:          memoryService,
	}
}
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	var req struct {
		Title     string `json:"title" binding:"required"`
		PersonaId string `json:"personaId" binding:"required"`
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
	existing, err := h.conversationRepository.GetConversationByPersonaAndUser(req.PersonaId, userId)
	if err == nil && existing != nil {
		res.ConversationId = existing.ID
		common.Success(c, res)
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	conversation := &model.Conversation{
		UserID:    userId,
		PersonaID: req.PersonaId,
		Title:     req.Title,
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

	userId, err := utils.GetUserIdFromSession(c)
	if err != nil {
		common.Fail(c, common.FailedCode)
		return
	}
	if persona.UserID != userId {
		common.Fail(c, common.FailedCode)
		return
	}

	conversation, err := h.conversationRepository.GetConversationById(req.ConversationId)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}

	if conversation.UserID != userId || (conversation.PersonaID != "" && conversation.PersonaID != req.PersonaId) {
		common.Fail(c, common.FailedCode)
		return
	}

	if conversation.PersonaID == "" {
		existing, err := h.conversationRepository.GetConversationByPersonaAndUser(req.PersonaId, userId)
		if err == nil && existing != nil && existing.ID != conversation.ID {
			common.Fail(c, common.FailedCode)
			return
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			common.Fail(c, common.DataBaseFailedCode)
			return
		}
		conversation.PersonaID = req.PersonaId
		if err := h.conversationRepository.UpdateConversation(conversation); err != nil {
			common.Fail(c, common.DataBaseFailedCode)
			return
		}
	}

	conversation_messages, err := h.conversationRepository.GetMessagesByConversationId(req.ConversationId)
	if err != nil {
		common.Fail(c, common.DataBaseFailedCode)
		return
	}
	conversation_messages = trimConversationRounds(conversation_messages, 30)

	// 构建增强的 System Prompt
	gsp := "回复时，你需要模拟微信聊天的回复风格，人们通常不会说完一大段话，而是一小段一小段的发送，请根据上下文和需求，合理分割回复内容，以\n分割。比如早啊，今天又是忙碌的一天。学生们要考地理生物，我还得布置考场，想想就头疼。你那边怎么样？，你需要以\n分割。早啊\n今天又是忙碌的一天n学生们要考地理生物\n我还得布置考场\n想想就头疼\n你那边怎么样？"
	enhancedSystemPrompt := persona.SystemPrompt + gsp
	enhancedSystemPrompt += "\n\n当前 personaId: " + req.PersonaId + "\n如需检索记忆，请调用 RetrieveMemories 工具，并填写 query。"

	tools := make([]tool.BaseTool, 0, 1)
	memoryTool, err := llm_tools.NewRetrieveMemoriesTool(h.memoryService, req.PersonaId, userId)
	if err != nil {
		utils.Log.Warn("创建记忆检索工具失败", zap.Error(err))
	} else {
		tools = append(tools, memoryTool)
	}
	resp, err := chat_core.Chat(c, req.Query, conversation_messages, enhancedSystemPrompt, tools...)
	res.Message = resp
	if err != nil {
		utils.Log.Error("聊天失败", zap.Error(err))
		common.Fail(c, common.ChatFailedCode)
		return
	}

	// 异步累积消息用于记忆提取
	go h.memoryService.AccumulateMessage(
		context.Background(),
		req.ConversationId,
		req.PersonaId,
		userId,
		req.Query,
		resp,
	)

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

func trimConversationRounds(messages []model.Message, maxRounds int) []model.Message {
	if maxRounds <= 0 {
		return []model.Message{}
	}
	limit := maxRounds * 2
	if len(messages) <= limit {
		return messages
	}
	return messages[len(messages)-limit:]
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
