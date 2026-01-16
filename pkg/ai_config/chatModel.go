package ai_config

type DeepSeekChatModel struct {
	Model   string `json:"model"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

var DeepSeekChatConfig DeepSeekChatModel
