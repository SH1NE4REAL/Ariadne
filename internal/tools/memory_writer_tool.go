package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ariadne/internal/model"
	"ariadne/internal/provider/embedding"
	"ariadne/internal/provider/vector"
)

const UserMemoryCollection = "ariadne_user_memory"

type MemoryWriterTool struct {
	Name              string
	Description       string
	EmbeddingProvider embedding.DashScopeEmbeddingProvider
	VectorClient      vector.ZillizRESTClient
}

func NewMemoryWriterTool() MemoryWriterTool {
	return MemoryWriterTool{
		Name:              "memory_writer_tool",
		Description:       "将高价值用户偏好写入 Zilliz 向量记忆库",
		EmbeddingProvider: embedding.NewDashScopeEmbeddingProvider(),
		VectorClient:      vector.NewZillizRESTClient(),
	}
}

func (t MemoryWriterTool) WriteUserMemory(
	ctx context.Context,
	req model.MemoryWriteRequest,
) (model.MemoryRecord, error) {
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		return model.MemoryRecord{}, fmt.Errorf("memory text is empty")
	}

	if req.UserID == "" {
		req.UserID = "anonymous"
	}

	if req.MemoryType == "" {
		req.MemoryType = "user_preference"
	}

	if req.Source == "" {
		req.Source = "user_message"
	}

	if req.Importance <= 0 {
		req.Importance = 5
	}

	vectorData, err := t.EmbeddingProvider.EmbedDocument(ctx, req.Text)
	if err != nil {
		return model.MemoryRecord{}, fmt.Errorf("embed memory document failed: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	id := fmt.Sprintf("mem_%s_%d", req.UserID, time.Now().UnixNano())

	entity := map[string]any{
		"id":            id,
		"vector":        vectorData,
		"user_id":       req.UserID,
		"session_id":    req.SessionID,
		"memory_type":   req.MemoryType,
		"text":          req.Text,
		"source":        req.Source,
		"importance":    req.Importance,
		"created_at":    now,
		"metadata_json": req.MetadataJSON,
	}

	err = t.VectorClient.Insert(ctx, UserMemoryCollection, []map[string]any{entity})
	if err != nil {
		return model.MemoryRecord{}, fmt.Errorf("insert memory into zilliz failed: %w", err)
	}

	return model.MemoryRecord{
		ID:           id,
		UserID:       req.UserID,
		SessionID:    req.SessionID,
		MemoryType:   req.MemoryType,
		Text:         req.Text,
		Source:       req.Source,
		Importance:   req.Importance,
		CreatedAt:    now,
		MetadataJSON: req.MetadataJSON,
	}, nil
}