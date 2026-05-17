package llm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"ariadne/internal/model"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

func GenerateTripSummaryWithLLM(ctx context.Context, plan model.FinalTripPlan, config model.LLMConfig) (string, error) {
	if config.APIKey == "" || config.Model == "" || config.BaseURL == "" {
		return "", errors.New("llm config is incomplete")
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  config.APIKey,
		Model:   config.Model,
		BaseURL: config.BaseURL,
	})
	if err != nil {
		return "", err
	}

	planBytes, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}

	resp, err := chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(buildTripSummarySystemPrompt()),
		schema.UserMessage(string(planBytes)),
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

func buildTripSummarySystemPrompt() string {
	return `你是 Ariadne 旅游规划 Agent 的总结助手。

你会收到一份 JSON 格式的旅行计划。

请根据其中的交通方案、景点、每日路线、预算和用户偏好，生成一段自然、简洁、有建议感的中文总结。

要求：
1. 只返回一段中文总结。
2. 不要返回 markdown。
3. 不要返回 JSON。
4. 不要编造 JSON 中没有的信息。
5. 语气像一个可靠的旅行规划助手。`
}