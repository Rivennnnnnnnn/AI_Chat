package utils

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // 建议用 UUID 作为 Key
)

func GenerateSessionId(username string) (string, error) {
	sessionId := uuid.New().String()
	return sessionId, nil
}

func GetUserIdFromSession(c *gin.Context) (int64, error) {
	userSession, exists := c.Get("userSession")
	if !exists {
		return 0, errors.New("user session not found")
	}
	userSessionMap, _ := userSession.(map[string]string)
	userId, err := strconv.ParseInt(userSessionMap["id"], 10, 64)
	if err != nil {
		return 0, errors.New("user id not found")
	}
	return userId, nil
}
