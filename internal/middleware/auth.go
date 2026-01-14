package middleware

import (
	"AI_Chat/internal/common"
	"AI_Chat/internal/repository"

	"github.com/gin-gonic/gin"
)

func Auth(userSessionRepository *repository.UserSessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionId := c.GetHeader("SessionId")
		if sessionId == "" {
			common.Fail(c, common.SessionExpiredCode)
			c.Abort()
			return
		}
		userSession, err := userSessionRepository.GetUserSession(sessionId, c)
		if err != nil {
			common.Fail(c, common.SessionExpiredCode)
			//fmt.Println(err)
			c.Abort()
			return
		}
		if len(userSession) == 0 {
			common.Fail(c, common.SessionExpiredCode)
			//fmt.Println("userSession is empty")
			c.Abort()
			return
		}
		c.Set("userSession", userSession)
		c.Next()
	}
}
