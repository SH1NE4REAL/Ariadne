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
	transportPlans := a.TransportTool.Run(request)
	attractions := a.AttractionTool.Run(request)
	dailyRoutes := a.RouteTool.Run(request, attractions)

	totalCost := calculateTotalCost(transportPlans, dailyRoutes)
	summary := generateSummary(request, totalCost)

	return model.FinalTripPlan{
		Request:            request,
		TransportPlans:     transportPlans,
		Attractions:        attractions,
		DailyRoutes:        dailyRoutes,
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