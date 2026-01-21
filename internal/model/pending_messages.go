package model

import "time"

// PendingMessages 待处理的对话消息（存储在 Redis）
type PendingMessages struct {
	ConversationID string        `json:"conversation_id"`
	PersonaID      string        `json:"persona_id"`
	UserID         int64         `json:"user_id"`
	RoundCount     int           `json:"round_count"`
	Messages       []MessagePair `json:"messages"`
}

// MessagePair 一轮对话的消息对
type MessagePair struct {
	UserMsg      string    `json:"user_msg"`
	AssistantMsg string    `json:"assistant_msg"`
	Timestamp    time.Time `json:"timestamp"`
}
