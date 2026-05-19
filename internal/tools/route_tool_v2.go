package tools

import "ariadne/internal/model"

type RouteTool struct {
	Name        string
	Description string
}

func NewRouteTool() RouteTool {
	return RouteTool{
		Name:        "route_tool",
		Description: "根据旅行请求和景点列表生成每日行程路线",
	}
}

func (t RouteTool) Run(request model.TripRequest, attractions []model.Attraction) []model.DailyRoute {
	return GenerateDailyRoutes(request, attractions)
}

func (t RouteTool) RunWithPreferences(
	request model.TripRequest,
	attractions []model.Attraction,
	preferenceProfile model.EffectivePreferenceProfile,
) []model.DailyRoute {
	return ComposeDailyRoutes(request, attractions, preferenceProfile)
}
