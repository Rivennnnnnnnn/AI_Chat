package repository

import (
	"AI_Chat/internal/model"
	"AI_Chat/pkg/utils"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type UserSessionRepository struct {
	redis *redis.Client
}

func NewUserSessionRepository(redis *redis.Client) *UserSessionRepository {
	return &UserSessionRepository{redis: redis}
}
func (r *UserSessionRepository) SetUserSession(user_base *model.UserBase, ctx *gin.Context) (string, error) {
	sessionId, err := utils.GenerateSessionId(user_base.Username)
	if err != nil {
		return "", err
	}
	inputs := fmt.Sprintf("%s:%s", "sessionId", sessionId)
	err = r.redis.HSet(ctx, inputs, user_base).Err()
	if err != nil {
		return "", err
	}
	err = r.redis.Expire(ctx, inputs, 24*time.Hour).Err()
	if err != nil {
		return "", err
	}
	return inputs, nil
}
func (r *UserSessionRepository) GetUserSession(sessionId string, ctx *gin.Context) (map[string]string, error) {
	val, err := r.redis.HGetAll(ctx, sessionId).Result()
	if err == redis.Nil {
		// 情况 A: Key 已过期，或者这个 Field 不存在
		return val, nil // 或者根据你的业务逻辑返回自定义的 "Session Expired" 错误
	} else if err != nil {
		// 情况 B: 真正的错误（如 Redis 断开连接、超时等）
		return val, err
	}

	// 情况 C: 成功获取到数据
	return val, nil
}

func (r *UserSessionRepository) DeleteUserSession(sessionId string, ctx *gin.Context) error {
	return r.redis.Del(ctx, sessionId).Err()
}
