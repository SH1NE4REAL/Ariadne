package tools

import "ariadne/internal/model"

type TransportTool struct {
	Name        string
	Description string
}

func NewTransportTool() TransportTool {
	return TransportTool{
		Name:        "transport_tool",
		Description: "根据出发地、目的地、预算和偏好生成交通方案",
	}
}

func (t TransportTool) Run(request model.TripRequest) []model.TransportPlan {
	return GenerateTransportPlans(request)
}