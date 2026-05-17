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

请根据其中的用户需求、最优交通方案、预算拆分、住宿建议、景点、每日路线和总预估费用，生成一段自然、简洁、有建议感的中文总结。

重要说明：
1. total_estimated_cost 表示当前系统估算的总费用，通常包括推荐交通费用、推荐住宿费用和景点基础费用。
2. best_booking_option 是当前推荐的交通购买或查询方案。
3. hotel_options 是当前推荐的住宿档位和预算建议。
4. budget_breakdown 是预算拆分建议。
5. 不要声称 total_estimated_cost 不包含交通或住宿，除非 JSON 明确说明。
6. 不要编造 JSON 中没有的信息，例如真实车次、真实酒店名、真实票价、真实余票。
7. 如果地图经纬度为空或地理编码失败，不要在总结中强调路线距离或地图精确性。

输出要求：
1. 只返回一段中文总结。
2. 不要返回 markdown。
3. 不要返回 JSON。
4. 语气像一个可靠的旅行规划助手。`
}