package llm

import (
	"context"
	"encoding/json"
	"errors"

	"ariadne/internal/model"
)

func GenerateTripSummaryWithLLM(ctx context.Context, plan model.FinalTripPlan, client *LLMClient) (string, error) {
	if client == nil {
		return "", errors.New("llm client is nil")
	}

	planBytes, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}

	return client.Generate(ctx, buildTripSummarySystemPrompt(), string(planBytes))
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