package handler

import (
	"AI_Chat/internal/common"
	"AI_Chat/internal/model"
	"AI_Chat/internal/repository"
	"AI_Chat/pkg/utils"
	_ "net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	userBaseRepository    *repository.UserBaseRepository
	userSessionRepository *repository.UserSessionRepository
}

func NewAuthHandler(userBaseRepository *repository.UserBaseRepository, userSessionRepository *repository.UserSessionRepository) *AuthHandler {
	return &AuthHandler{userBaseRepository: userBaseRepository, userSessionRepository: userSessionRepository}
}
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Password string `json:"password" binding:"required,min=8,max=20"`
		Email    string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.FailedCode)
		return
	}
	password, err := utils.HashEncode(req.Password)
	if err != nil {
		utils.Log.Error("密码加密失败", zap.Error(err))
		common.Fail(c, common.FailedCode)
		return
	}
	err = h.userBaseRepository.CreateUserBase(&model.UserBase{
		Username: req.Username,
		Password: password,
		Email:    req.Email,
	})
	if err != nil {
		common.Fail(c, common.RegisterFailedCode)
		return
	}
	common.Success(c, nil)
}
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Password string `json:"password" binding:"required,min=8,max=20"`
	}
	var res struct {
		SessionId string `json:"sessionId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.FailedCode)
		return
	}
	userBase, err := h.userBaseRepository.GetUserBaseByUsername(req.Username)
	if err != nil {
		common.Fail(c, common.FailedCode)
		return
	}
	if userBase == nil {
		common.Fail(c, common.LoginFailedCode)
		return
	}
	if !utils.HashCompare(req.Password, userBase.Password) {
		common.Fail(c, common.LoginFailedCode)
		return
	}
	res.SessionId, err = h.userSessionRepository.SetUserSession(userBase, c)
	if err != nil {
		common.Fail(c, common.RedisFailedCode)
		return
	}
	common.Success(c, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionId := c.GetHeader("SessionId")
	if sessionId == "" {
		common.Fail(c, common.FailedCode)
		return
	}
	err := h.userSessionRepository.DeleteUserSession(sessionId, c)
	if err != nil {
		common.Fail(c, common.RedisFailedCode)
		return
	}
	common.Success(c, nil)
}
