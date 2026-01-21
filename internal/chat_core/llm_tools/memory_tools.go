package llm_tools

import (
	"context"
	"fmt"
	"strings"

	"AI_Chat/internal/memory"
	"AI_Chat/pkg/utils"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"go.uber.org/zap"
)

type RetrieveMemoriesParams struct {
	PersonaID string `json:"personaId" jsonschema:"当前对话的personaId"`
	Query     string `json:"query" jsonschema:"用于检索记忆的查询内容"`
	TopK      int    `json:"topK,omitempty" jsonschema:"返回的记忆数量，默认5"`
}

func NewRetrieveMemoriesTool(memoryService *memory.MemoryService, defaultPersonaID string, userId int64) (tool.InvokableTool, error) {
	return toolutils.InferTool(
		"RetrieveMemories",
		"根据personaId和query检索相关记忆，返回格式化结果",
		func(ctx context.Context, params *RetrieveMemoriesParams) (string, error) {
			if memoryService == nil {
				return "", fmt.Errorf("memory service is nil")
			}
			if params == nil {
				return "", fmt.Errorf("params is nil")
			}
			personaID := strings.TrimSpace(params.PersonaID)
			if personaID == "" {
				personaID = strings.TrimSpace(defaultPersonaID)
			}
			if personaID == "" {
				return "", fmt.Errorf("personaId is required")
			}
			topK := params.TopK
			if topK <= 0 {
				topK = 5
			}
			utils.Log.Info("记忆检索输入",
				zap.String("personaId", personaID),
				zap.Int64("userId", userId),
				zap.Int("topK", topK),
				zap.String("query", params.Query),
			)
			memories, err := memoryService.RetrieveMemories(ctx, personaID, userId, params.Query, topK)
			if err != nil {
				return "", err
			}
			formatted := strings.TrimSpace(memoryService.FormatMemoriesForPrompt(memories))
			utils.Log.Info("记忆检索结果",
				zap.String("personaId", personaID),
				zap.Int64("userId", userId),
				zap.Int("count", len(memories)),
				zap.String("formatted", formatted),
			)
			if formatted == "" {
				return "没有检索到相关记忆。", nil
			}
			return formatted, nil
		},
	)
}
