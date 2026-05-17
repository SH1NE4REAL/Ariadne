package llm

import (
	"context"
	"errors"
	"strings"

	"ariadne/internal/model"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

type LLMClient struct {
	chatModel *openai.ChatModel
}

func NewLLMClient(ctx context.Context, config model.LLMConfig) (*LLMClient, error) {
	if !IsLLMConfigComplete(config) {
		return nil, errors.New("llm config is incomplete")
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  config.APIKey,
		Model:   config.Model,
		BaseURL: config.BaseURL,
	})
	if err != nil {
		return nil, err
	}

	return &LLMClient{
		chatModel: chatModel,
	}, nil
}

func IsLLMConfigComplete(config model.LLMConfig) bool {
	return config.APIKey != "" && config.Model != "" && config.BaseURL != ""
}

func (c *LLMClient) Generate(ctx context.Context, systemPrompt string, userMessage string) (string, error) {
	if c == nil || c.chatModel == nil {
		return "", errors.New("llm client is nil")
	}

	resp, err := c.chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userMessage),
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}