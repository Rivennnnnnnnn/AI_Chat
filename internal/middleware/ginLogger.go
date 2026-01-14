package middleware

import (
	"AI_Chat/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 先执行后面的逻辑
		c.Next()

		// 请求结束后的统计
		cost := time.Since(start)
		status := c.Writer.Status()

		// 统一记录所有接口的访问情况
		utils.Log.Info(path,
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("cost", cost),
		)
	}
}
