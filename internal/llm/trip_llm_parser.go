package llm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"ariadne/internal/model"
)

func ParseTripRequestWithLLM(ctx context.Context, message string, client *LLMClient) (model.TripRequest, error) {
	if client == nil {
		return model.TripRequest{}, errors.New("llm client is nil")
	}

	content, err := client.Generate(ctx, buildTripParserSystemPrompt(), message)
	if err != nil {
		return model.TripRequest{}, err
	}

	jsonText := extractJSON(content)

	var tripRequest model.TripRequest
	err = json.Unmarshal([]byte(jsonText), &tripRequest)
	if err != nil {
		return model.TripRequest{}, err
	}

	tripRequest.RawInput = message

	return tripRequest, nil
}

func buildTripParserSystemPrompt() string {
	return `你是 Ariadne 旅游规划 Agent 的意图解析器。

你的任务是把用户的自然语言旅行需求解析成 JSON。

必须只返回 JSON，不要返回 markdown，不要解释，不要代码块。

JSON 字段如下：
{
  "origin": "出发地，没有则为空字符串",
  "destination": "目的地，没有则为空字符串",
  "days": 天数，没有则为0,
  "budget": 预算数字，没有则为0,
  "preference": "用户偏好，例如轻松、省钱、美食、拍照，没有则为空字符串"
}

规则：
1. 如果用户说“我人在南京”“南京出发”“从南京开始”，origin 都应该是“南京”。
2. 如果用户说“去杭州”“目的地杭州”“想玩杭州”，destination 都应该是“杭州”。
3. 如果用户说“三天”，days 应该是 3。
4. 如果用户说“预算1500”“大概花1500”，budget 应该是 1500。
5. 只返回 JSON。`
}

func extractJSON (text string) string {
	text = strings.TrimSpace(text)

	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")

	return strings.TrimSpace(text)
}