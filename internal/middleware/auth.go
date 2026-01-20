package middleware

import (
	"AI_Chat/internal/common"
	"AI_Chat/internal/repository"
	"AI_Chat/pkg/utils"

	"github.com/gin-gonic/gin"
)

func Auth(userSessionRepository *repository.UserSessionRepository, userBaseRepository *repository.UserBaseRepository) gin.HandlerFunc {
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

		//检查用户是否存在
		userId, err := utils.GetUserIdFromSession(c)
		if err != nil {
			common.Fail(c, common.FailedCode)
			c.Abort()
			return
		}
		user, err := userBaseRepository.FindByID(userId)
		if err != nil || user == nil {
			common.Fail(c, common.UserNotFoundCode)
			c.Abort()
			return
		}

		c.Next()
	}
}
