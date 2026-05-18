package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"ariadne/internal/model"
	"ariadne/internal/provider/embedding"
	"ariadne/internal/provider/vector"
)

type MemoryRetrieverTool struct {
	Name              string
	Description       string
	EmbeddingProvider embedding.DashScopeEmbeddingProvider
	VectorClient      vector.ZillizRESTClient
}

func NewMemoryRetrieverTool() MemoryRetrieverTool {
	return MemoryRetrieverTool{
		Name:              "memory_retriever_tool",
		Description:       "从 Zilliz 向量记忆库检索与当前用户输入相关的长期记忆",
		EmbeddingProvider: embedding.NewDashScopeEmbeddingProvider(),
		VectorClient:      vector.NewZillizRESTClient(),
	}
}

func (t MemoryRetrieverTool) RetrieveUserMemories(
	ctx context.Context,
	userID string,
	query string,
	limit int,
) ([]model.MemoryRecord, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("memory query is empty")
	}

	if userID == "" {
		userID = "anonymous"
	}

	if limit <= 0 {
		limit = 5
	}

	queryVector, err := t.EmbeddingProvider.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed memory query failed: %w", err)
	}

	filter := fmt.Sprintf(`user_id == "%s"`, escapeFilterString(userID))

	outputFields := []string{
		"id",
		"user_id",
		"session_id",
		"memory_type",
		"text",
		"source",
		"importance",
		"created_at",
		"metadata_json",
	}

	rows, err := t.VectorClient.Search(
		ctx,
		UserMemoryCollection,
		queryVector,
		limit,
		filter,
		outputFields,
	)
	if err != nil {
		return nil, fmt.Errorf("search user memories failed: %w", err)
	}

	memories := make([]model.MemoryRecord, 0, len(rows))

	for _, row := range rows {
		memories = append(memories, parseMemorySearchRow(row))
	}

	return memories, nil
}

func parseMemorySearchRow(row map[string]any) model.MemoryRecord {
	entity := row

	if rawEntity, ok := row["entity"].(map[string]any); ok {
		entity = rawEntity
	}

	return model.MemoryRecord{
		ID:           getString(entity, "id"),
		UserID:       getString(entity, "user_id"),
		SessionID:    getString(entity, "session_id"),
		MemoryType:   getString(entity, "memory_type"),
		Text:         getString(entity, "text"),
		Source:       getString(entity, "source"),
		Importance:   getInt(entity, "importance"),
		CreatedAt:    getString(entity, "created_at"),
		MetadataJSON: getString(entity, "metadata_json"),
		Score:        getFloat(row, "distance"),
	}
}

func getString(m map[string]any, key string) string {
	value, ok := m[key]
	if !ok || value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

func getInt(m map[string]any, key string) int {
	value, ok := m[key]
	if !ok || value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		return 0
	}
}

func getFloat(m map[string]any, key string) float64 {
	value, ok := m[key]
	if !ok || value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case string:
		n, _ := strconv.ParseFloat(v, 64)
		return n
	default:
		return 0
	}
}

func escapeFilterString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}