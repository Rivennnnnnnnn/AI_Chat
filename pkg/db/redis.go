package db

import (
	"AI_Chat/pkg/utils"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9" // 如果用 postgres 就换成 gorm.io/driver/postgres
)

var RedisClient *redis.Client
var ctx = context.Background()

func InitRedis(cfg utils.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	_, err := RedisClient.Ping(ctx).Result()
	return err
}
