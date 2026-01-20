package chat_core

import (
	_ "AI_Chat/internal/chat_core/llm_tools"
	"AI_Chat/internal/model"
	"AI_Chat/pkg/ai_config"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/prompt"
	_ "github.com/cloudwego/eino/components/tool"
	_ "github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

func Chat(c *gin.Context, query string, history []model.Message, system_prompt string) (string, error) {
	cm, err := deepseek.NewChatModel(c, &deepseek.ChatModelConfig{
		APIKey:  ai_config.DeepSeekChatConfig.APIKey,
		Model:   ai_config.DeepSeekChatConfig.Model,
		BaseURL: ai_config.DeepSeekChatConfig.BaseURL,
	})
	// mytool, err := llm_tools.NewGetSecretTool()
	// tools := compose.ToolsNodeConfig{
	// 	Tools: []tool.BaseTool{mytool},
	// }
	agent, err := react.NewAgent(c, &react.AgentConfig{
		ToolCallingModel: cm,
		MaxStep:          5,
	})
	if err != nil {
		return "Agent创建出错", err
	}
	// sort.Slice(history, func(i, j int) bool {
	// 	return history[i].CreatedAt.Before(history[j].CreatedAt)
	// })
	//带历史记录的聊天
	historyMessages := make([]*schema.Message, 0)
	for _, message := range history {
		switch message.Role {
		case "user":
			historyMessages = append(historyMessages, &schema.Message{
				Role:    schema.User,
				Content: message.Content,
			})
		case "assistant":
			historyMessages = append(historyMessages, &schema.Message{
				Role:    schema.Assistant,
				Content: message.Content,
			})
		case "system":
			historyMessages = append(historyMessages, &schema.Message{
				Role:    schema.System,
				Content: message.Content,
			})
		}
	}
	template := prompt.FromMessages(schema.FString,
		// 系统消息模板
		schema.SystemMessage(system_prompt),

		// 插入需要的对话历史（新对话的话这里不填）
		schema.MessagesPlaceholder("chat_history", true),

		// 用户消息模板
		schema.UserMessage(query),
	)
	messages, err := template.Format(c, map[string]any{
		"chat_history": historyMessages,
	})
	if err != nil {
		return "格式化出错，请检查格式", err
	}
	resp, err := agent.Generate(c, messages)
	if err != nil {
		return "生成出错，请检查配置文件或网络", err
	}
	return resp.Content, nil
}
