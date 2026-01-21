package memory

import (
	"context"
	"testing"
	"time"

	"AI_Chat/pkg/ai_config"
	"AI_Chat/pkg/utils"
)

func TestDeepSeekEmbedding(t *testing.T) {
	if err := utils.InitConfig(); err != nil {
		t.Fatalf("init config failed: %v", err)
	}
	cfg := ai_config.DeepSeekEmbeddingConfig
	if cfg.APIKey == "" || cfg.Model == "" {
		t.Skip("DeepSeek embedding config missing")
	}

	client := NewEmbeddingClient(cfg)
	if client == nil {
		t.Fatalf("embedding client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	vec, err := client.Embed(ctx, "测试 embedding 接口")
	if err != nil {
		t.Fatalf("embed failed: %v", err)
	}
	if len(vec) == 0 {
		t.Fatalf("embedding vector is empty")
	}
}
