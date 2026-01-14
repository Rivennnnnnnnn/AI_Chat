package handler

import (
	"AI_Chat/internal/common"

	"github.com/gin-gonic/gin"
)

type TestHandler struct {
}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}
func (h *TestHandler) Test(c *gin.Context) {
	common.Success(c, nil)
}