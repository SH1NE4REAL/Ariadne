package agent

import (
	"fmt"

	"ariadne/internal/model"
	"ariadne/internal/tools"
)

type TripAgent struct {
	TransportTool  tools.TransportTool
	AttractionTool tools.AttractionTool
	RouteTool      tools.RouteTool
}

func NewTripAgent() TripAgent {
	return TripAgent{
		TransportTool:  tools.NewTransportTool(),
		AttractionTool: tools.NewAttractionTool(),
		RouteTool:      tools.NewRouteTool(),
	}
}

func (a TripAgent) Plan(request model.TripRequest) model.FinalTripPlan {
	agentSteps := make([]model.AgentStep, 0)

	transportPlans := a.TransportTool.Run(request)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.TransportTool.Name,
		Description: "根据出发地、目的地、预算和偏好生成交通方案",
	})

	attractions := a.AttractionTool.Run(request)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.AttractionTool.Name,
		Description: "根据目的地、旅行天数和用户偏好推荐景点",
	})

	dailyRoutes := a.RouteTool.Run(request, attractions)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.RouteTool.Name,
		Description: "根据旅行请求和景点列表生成每日行程路线",
	})

	totalCost := calculateTotalCost(transportPlans, dailyRoutes)
	summary := generateSummary(request, totalCost)

	return model.FinalTripPlan{
		Request:            request,
		TransportPlans:     transportPlans,
		Attractions:        attractions,
		DailyRoutes:        dailyRoutes,
		AgentSteps:         agentSteps,
		TotalEstimatedCost: totalCost,
		Summary:            summary,
	}
}

func calculateTotalCost(transportPlans []model.TransportPlan, dailyRoutes []model.DailyRoute) int {
	total := 0

	if len(transportPlans) > 0 {
		total += transportPlans[0].Price
	}

	for _, route := range dailyRoutes {
		total += route.EstimatedCost
	}

	return total
}

func generateSummary(request model.TripRequest, totalCost int) string {
	if request.Budget > 0 && totalCost > request.Budget {
		return fmt.Sprintf("当前方案预估花费约 %d 元，可能超过你的预算 %d 元，建议减少高消费项目或选择更经济的交通方式。", totalCost, request.Budget)
	}

	if request.Preference == "轻松" {
		return fmt.Sprintf("当前方案预估花费约 %d 元，整体节奏较轻松，适合不想太赶的旅行。", totalCost)
	}

	if request.Preference == "省钱" {
		return fmt.Sprintf("当前方案预估花费约 %d 元，已尽量控制景点和交通成本。", totalCost)
	}

	return fmt.Sprintf("当前方案预估花费约 %d 元，可作为初版旅行计划参考。", totalCost)
}