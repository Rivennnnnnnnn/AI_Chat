package memory

import (
	"AI_Chat/internal/model"
	"AI_Chat/internal/repository"
	"AI_Chat/pkg/ai_config"
	"AI_Chat/pkg/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	PendingMessagesKeyPrefix = "chat:pending:"
	PendingMessagesTTL       = 24 * time.Hour
	ExtractThreshold         = 10 // 每10轮触发提取
)

type EmbeddingClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewEmbeddingClient(config ai_config.DeepSeekEmbeddingModel) *EmbeddingClient {
	if strings.TrimSpace(config.APIKey) == "" || strings.TrimSpace(config.Model) == "" {
		return nil
	}
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	return &EmbeddingClient{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		model:   config.Model,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *EmbeddingClient) Embed(ctx context.Context, input string) ([]float32, error) {
	if c == nil {
		return nil, fmt.Errorf("embedding client is nil")
	}
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("embedding input is empty")
	}
	reqBody := map[string]interface{}{
		"model":      c.model,
		"input":      input,
		"dimensions": 1536,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyText := strings.TrimSpace(string(bodyBytes))
		if len(bodyText) > 1000 {
			bodyText = bodyText[:1000] + "..."
		}
		return nil, fmt.Errorf("embedding request failed: status=%d url=%s model=%s body=%s", resp.StatusCode, c.baseURL+"/v1/embeddings", c.model, bodyText)
	}

	var parsed struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) == 0 || len(parsed.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("embedding response empty")
	}

	result := make([]float32, 0, len(parsed.Data[0].Embedding))
	for _, v := range parsed.Data[0].Embedding {
		result = append(result, float32(v))
	}
	return result, nil
}

type MemoryService struct {
	memoryRepo  *repository.MemoryRepository
	redisClient *redis.Client
	embedding   *EmbeddingClient
	milvusStore *MilvusStore
}

func NewMemoryService(memoryRepo *repository.MemoryRepository, redisClient *redis.Client, milvusStore *MilvusStore) *MemoryService {
	return &MemoryService{
		memoryRepo:  memoryRepo,
		redisClient: redisClient,
		embedding:   NewEmbeddingClient(ai_config.DeepSeekEmbeddingConfig),
		milvusStore: milvusStore,
	}
}

// AccumulateMessage 累积消息，检查是否触发提取
func (s *MemoryService) AccumulateMessage(ctx context.Context, convId, personaId string, userId int64, userMsg, aiReply string) error {
	key := PendingMessagesKeyPrefix + convId

	// 获取现有的待处理消息
	pending, err := s.getPendingMessages(ctx, key)
	if err != nil && err != redis.Nil {
		return err
	}

	if pending == nil {
		pending = &model.PendingMessages{
			ConversationID: convId,
			PersonaID:      personaId,
			UserID:         userId,
			RoundCount:     0,
			Messages:       []model.MessagePair{},
		}
	}

	// 追加消息
	pending.Messages = append(pending.Messages, model.MessagePair{
		UserMsg:      userMsg,
		AssistantMsg: aiReply,
		Timestamp:    time.Now(),
	})
	pending.RoundCount++

	// 检查是否达到阈值
	if pending.RoundCount >= ExtractThreshold {
		// 触发提取（异步）
		go s.ExtractMemoriesFromPending(context.Background(), pending)
		// 清空缓存
		return s.clearPendingMessages(ctx, key)
	}

	// 保存到 Redis
	return s.savePendingMessages(ctx, key, pending)
}

// getPendingMessages 从 Redis 获取待处理消息
func (s *MemoryService) getPendingMessages(ctx context.Context, key string) (*model.PendingMessages, error) {
	data, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var pending model.PendingMessages
	if err := json.Unmarshal(data, &pending); err != nil {
		return nil, err
	}
	return &pending, nil
}

// savePendingMessages 保存待处理消息到 Redis
func (s *MemoryService) savePendingMessages(ctx context.Context, key string, pending *model.PendingMessages) error {
	data, err := json.Marshal(pending)
	if err != nil {
		return err
	}
	return s.redisClient.Set(ctx, key, data, PendingMessagesTTL).Err()
}

// clearPendingMessages 清空待处理消息
func (s *MemoryService) clearPendingMessages(ctx context.Context, key string) error {
	return s.redisClient.Del(ctx, key).Err()
}

// ExtractMemoriesFromPending 从待处理消息中提取记忆
func (s *MemoryService) ExtractMemoriesFromPending(ctx context.Context, pending *model.PendingMessages) error {
	if len(pending.Messages) == 0 {
		return nil
	}

	// 1. 构建对话文本
	conversationText := s.buildConversationText(pending.Messages)

	// 2. 获取现有记忆，用于冲突检测
	existingMemories, err := s.memoryRepo.GetActiveMemoriesByPersonaAndUser(pending.PersonaID, pending.UserID)
	if err != nil {
		fmt.Printf("获取现有记忆失败: %v\n", err)
	}

	// 3. 调用 LLM 提取并合并记忆
	newMemories, err := s.callLLMExtractAndMergeMemories(ctx, conversationText, existingMemories)
	if err != nil {
		fmt.Printf("提取并合并记忆失败: %v\n", err)
		return err
	}

	// 4. 处理提取结果
	for _, item := range newMemories {
		if item.Action == "add" {
			// 新增记忆
			memory := &model.Memory{
				PersonaID: pending.PersonaID,
				UserID:    pending.UserID,
				Type:      item.Type,
				Content:   item.Content,
				Keywords:  item.Keywords,
				Source:    model.MemorySourceAuto,
				Status:    model.MemoryStatusActive,
			}
			s.PrepareMemoryEmbedding(ctx, memory)
			if err := s.memoryRepo.CreateMemory(memory); err != nil {
				fmt.Printf("存储新增记忆失败: %v\n", err)
			} else {
				s.UpsertMilvusEmbedding(ctx, memory)
			}
		} else if item.Action == "update" && item.OldMemoryID != "" {
			// 更新记忆（合并冲突）
			// 1. 标记旧记忆为 superseded
			newMemory := &model.Memory{
				PersonaID: pending.PersonaID,
				UserID:    pending.UserID,
				Type:      item.Type,
				Content:   item.Content,
				Keywords:  item.Keywords,
				Source:    model.MemorySourceAuto,
				Status:    model.MemoryStatusActive,
			}
			s.PrepareMemoryEmbedding(ctx, newMemory)
			if err := s.memoryRepo.CreateMemory(newMemory); err != nil {
				fmt.Printf("存储更新后的记忆失败: %v\n", err)
				continue
			}
			s.UpsertMilvusEmbedding(ctx, newMemory)
			if err := s.memoryRepo.UpdateMemoryStatus(item.OldMemoryID, model.MemoryStatusSuperseded, newMemory.ID); err != nil {
				fmt.Printf("更新旧记忆状态失败: %v\n", err)
			}
		}
	}

	return nil
}

// buildConversationText 构建对话文本
func (s *MemoryService) buildConversationText(messages []model.MessagePair) string {
	text := ""
	for i, msg := range messages {
		text += fmt.Sprintf("--- 第%d轮 ---\n用户: %s\nAI: %s\n\n", i+1, msg.UserMsg, msg.AssistantMsg)
	}
	return text
}

type MemoryAction struct {
	Action      string `json:"action"` // add, update, none
	OldMemoryID string `json:"old_memory_id,omitempty"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Keywords    string `json:"keywords"`
}

// callLLMExtractAndMergeMemories 调用 LLM 提取并处理冲突
func (s *MemoryService) callLLMExtractAndMergeMemories(ctx context.Context, conversationText string, existing []model.Memory) ([]MemoryAction, error) {
	cm, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  ai_config.DeepSeekChatConfig.APIKey,
		Model:   ai_config.DeepSeekChatConfig.Model,
		BaseURL: ai_config.DeepSeekChatConfig.BaseURL,
	})
	if err != nil {
		return nil, err
	}

	existingText := ""
	for _, m := range existing {
		existingText += fmt.Sprintf("- ID: %s, 内容: %s\n", m.ID, m.Content)
	}

	systemPrompt := `你是记忆管理助手。请分析对话并更新用户记忆。

【任务】
1. 从【新对话】中提取关于用户的重要、持久、客观的信息。
2. 对比【现有记忆】，判断新信息是：
   - 新增 (add)：现有记忆中没有相关信息。
   - 更新 (update)：新信息与某条现有记忆相关且存在冲突或补充（需合并）。
   - 无需操作 (none)：新信息已存在或无价值。
3. 如果是更新，请生成合并后的精炼内容，并提供对应的 old_memory_id。

输出格式 JSON:
{
  "actions": [
    {"action": "add", "type": "fact", "content": "...", "keywords": "..."},
    {"action": "update", "old_memory_id": "mem:xxx", "type": "preference", "content": "合并后的内容", "keywords": "..."}
  ]
}`

	userPrompt := fmt.Sprintf("【现有记忆】\n%s\n\n【新对话】\n%s", existingText, conversationText)

	messages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: userPrompt},
	}

	resp, err := cm.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	content := resp.Content
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result struct {
		Actions []MemoryAction `json:"actions"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("解析记忆Action失败: %v, content: %s", err, content)
	}

	return result.Actions, nil
}

func (s *MemoryService) PrepareMemoryEmbedding(ctx context.Context, memory *model.Memory) {
	if s.embedding == nil || memory == nil {
		return
	}
	embedding, err := s.embedding.Embed(ctx, memory.Content)
	if err != nil {
		return
	}
	embeddingText, err := marshalEmbedding(embedding)
	if err != nil {
		return
	}
	now := time.Now()
	memory.Embedding = embeddingText
	memory.EmbeddingUpdatedAt = &now
}

func (s *MemoryService) EnsureMemoryEmbedding(ctx context.Context, memory *model.Memory) {
	if s.embedding == nil || memory == nil || memory.Embedding != "" {
		return
	}
	embedding, err := s.embedding.Embed(ctx, memory.Content)
	if err != nil {
		return
	}
	embeddingText, err := marshalEmbedding(embedding)
	if err != nil {
		return
	}
	if err := s.memoryRepo.UpdateMemoryEmbedding(memory.ID, embeddingText); err != nil {
		return
	}
	memory.Embedding = embeddingText
	now := time.Now()
	memory.EmbeddingUpdatedAt = &now
}

func (s *MemoryService) UpsertMilvusEmbedding(ctx context.Context, memory *model.Memory) {
	if s.milvusStore == nil || memory == nil || memory.ID == "" {
		return
	}
	vec, err := parseEmbedding(memory.Embedding)
	if err != nil || len(vec) == 0 {
		if s.embedding == nil {
			return
		}
		embedding, err := s.embedding.Embed(ctx, memory.Content)
		if err != nil {
			return
		}
		memory.Embedding, err = marshalEmbedding(embedding)
		if err != nil {
			return
		}
		vec = embedding
		now := time.Now()
		memory.EmbeddingUpdatedAt = &now
		_ = s.memoryRepo.UpdateMemoryEmbedding(memory.ID, memory.Embedding)
	}
	_ = s.milvusStore.UpsertMemory(ctx, memory.ID, memory.PersonaID, memory.UserID, vec)
}

func (s *MemoryService) DeleteMilvusMemory(ctx context.Context, memoryID string) {
	if s.milvusStore == nil || memoryID == "" {
		return
	}
	_ = s.milvusStore.DeleteMemory(ctx, memoryID)
}

// RetrieveMemories 检索相关记忆（用于聊天时注入）
func (s *MemoryService) RetrieveMemories(ctx context.Context, personaId string, userId int64, query string, topK int) ([]model.Memory, error) {
	if s.embedding == nil || strings.TrimSpace(query) == "" || topK <= 0 {
		utils.Log.Info("记忆检索来源",
			zap.String("source", "db"),
			zap.String("reason", "embedding_disabled_or_query_empty"),
			zap.String("personaId", personaId),
			zap.Int64("userId", userId),
			zap.Int("topK", topK),
		)
		return s.memoryRepo.GetActiveMemoriesByPersonaAndUser(personaId, userId)
	}

	queryEmbedding, err := s.embedding.Embed(ctx, query)
	if err != nil {
		utils.Log.Info("记忆检索来源",
			zap.String("source", "db"),
			zap.String("reason", "embed_failed"),
			zap.String("personaId", personaId),
			zap.Int64("userId", userId),
			zap.Int("topK", topK),
		)
		return s.memoryRepo.GetActiveMemoriesByPersonaAndUser(personaId, userId)
	}

	if s.milvusStore != nil {
		ids, err := s.milvusStore.SearchMemories(ctx, personaId, userId, queryEmbedding, topK)
		if err == nil && len(ids) > 0 {
			records, err := s.memoryRepo.GetMemoriesByIDs(personaId, userId, ids)
			if err == nil && len(records) > 0 {
				byID := make(map[string]model.Memory, len(records))
				for _, mem := range records {
					byID[mem.ID] = mem
				}
				ordered := make([]model.Memory, 0, len(ids))
				for _, id := range ids {
					if mem, ok := byID[id]; ok {
						ordered = append(ordered, mem)
					}
				}
				if len(ordered) > 0 {
					utils.Log.Info("记忆检索来源",
						zap.String("source", "milvus"),
						zap.String("personaId", personaId),
						zap.Int64("userId", userId),
						zap.Int("topK", topK),
						zap.Int("count", len(ordered)),
					)
					return ordered, nil
				}
			}
		}
	}

	utils.Log.Info("记忆检索来源",
		zap.String("source", "db"),
		zap.String("reason", "milvus_miss_or_unavailable"),
		zap.String("personaId", personaId),
		zap.Int64("userId", userId),
		zap.Int("topK", topK),
	)
	memories, err := s.memoryRepo.GetActiveMemoriesByPersonaAndUser(personaId, userId)
	if err != nil {
		return nil, err
	}

	type scoredMemory struct {
		memory model.Memory
		score  float32
	}
	scored := make([]scoredMemory, 0, len(memories))
	for _, mem := range memories {
		if mem.Embedding == "" {
			s.EnsureMemoryEmbedding(ctx, &mem)
		}
		vec, err := parseEmbedding(mem.Embedding)
		if err != nil || len(vec) == 0 {
			continue
		}
		score := cosineSimilarity(queryEmbedding, vec)
		scored = append(scored, scoredMemory{memory: mem, score: score})
	}

	if len(scored) == 0 {
		return memories, nil
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	if topK > len(scored) {
		topK = len(scored)
	}
	result := make([]model.Memory, 0, topK)
	for i := 0; i < topK; i++ {
		result = append(result, scored[i].memory)
	}
	return result, nil
}

// FormatMemoriesForPrompt 格式化记忆用于注入 Prompt
func (s *MemoryService) FormatMemoriesForPrompt(memories []model.Memory) string {
	if len(memories) == 0 {
		return ""
	}

	result := "## 你对用户的了解：\n"
	for _, mem := range memories {
		result += fmt.Sprintf("- %s\n", mem.Content)
	}
	return result
}

func marshalEmbedding(vec []float32) (string, error) {
	if len(vec) == 0 {
		return "", fmt.Errorf("empty embedding")
	}
	data, err := json.Marshal(vec)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseEmbedding(data string) ([]float32, error) {
	if strings.TrimSpace(data) == "" {
		return nil, fmt.Errorf("empty embedding")
	}
	var vec []float32
	if err := json.Unmarshal([]byte(data), &vec); err != nil {
		return nil, err
	}
	return vec, nil
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	var dot, normA, normB float64
	for i := 0; i < limit; i++ {
		dot += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}
