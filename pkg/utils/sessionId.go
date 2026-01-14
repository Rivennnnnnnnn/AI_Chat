package utils

import (
	"github.com/google/uuid" // 建议用 UUID 作为 Key
)

func GenerateSessionId(username string) (string, error) {
	sessionId := uuid.New().String()
	return sessionId, nil
}
