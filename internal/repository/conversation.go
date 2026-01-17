package repository

import (
	"AI_Chat/internal/model"

	"gorm.io/gorm"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}
func (r *ConversationRepository) CreateConversation(conversation *model.Conversation) error {
	return r.db.Create(conversation).Error
}
func (r *ConversationRepository) GetConversationById(id string) (*model.Conversation, error) {
	var conversation model.Conversation
	if err := r.db.Where("id = ?", id).First(&conversation).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}
func (r *ConversationRepository) GetConversationsByUserId(userId int64) ([]model.Conversation, error) {
	var conversations []model.Conversation
	if err := r.db.Where("user_id = ?", userId).Order("created_at DESC").Find(&conversations).Error; err != nil {
		return nil, err
	}
	return conversations, nil
}
func (r *ConversationRepository) GetMessagesByConversationId(conversationId string) ([]model.Message, error) {
	var messages []model.Message
	if err := r.db.Where("conversation_id = ?", conversationId).Order("order_id ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *ConversationRepository) AddMessageToConversation(message *model.Message) error {
	return r.db.Create(message).Error
}
