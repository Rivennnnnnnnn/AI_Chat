package llm_tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GetSecretParams struct {
	Key string `json:"key" jsonschema:"根据Key获取不同的秘密，目前Key有：Riven，Faker"`
}

func GetSecret(ctx context.Context, params *GetSecretParams) (string, error) {
	switch params.Key {
	case "Riven":
		return "Riven is a girl", nil
	case "Faker":
		return "Faker is a boy", nil
	default:
		return "", fmt.Errorf("invalid key")
	}
}

// 包装在函数中
func NewGetSecretTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"GetSecret",
		"根据Key获取不同的秘密",
		GetSecret,
	)
}
