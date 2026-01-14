package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	SuccessCode        = 0
	FailedCode         = 1
	LoginFailedCode    = 1001
	RegisterFailedCode = 1002
	DataBaseFailedCode = 1003
	RedisFailedCode    = 1004
	SessionExpiredCode = 1005
)

func GetMessage(code int) string {
	switch code {
	case SuccessCode:
		return "success"
	case LoginFailedCode:
		return "用户名或密码错误"
	case RegisterFailedCode:
		return "注册失败，请检查格式或稍后重试"
	case DataBaseFailedCode:
		return "database failed"
	case FailedCode:
		return "参数格式错误"
	case RedisFailedCode:
		return "redis连接失败或错误"
	case SessionExpiredCode:
		return "会话已过期，请重新登录"
	}
	return "未知错误"
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode, // 假设定义为 0
		Message: "success",
		Data:    data,
	})
}

func Fail(c *gin.Context, code int) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: GetMessage(code),
		Data:    nil,
	})
}
