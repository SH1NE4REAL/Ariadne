package tools

import "ariadne/internal/model"

func GenerateDailyRoutes(request model.TripRequest, attractions []model.Attraction) []model.DailyRoute {
	if request.Days <= 0 {
		request.Days = 1
	}

	routes := make([]model.DailyRoute, 0)

	for day := 1; day <= request.Days; day++ {
		route := model.DailyRoute{
			Day:         day,
			Title:       generateRouteTitle(day, request),
			Attractions: pickAttractionsForDay(day, request.Days, attractions),
			Summary:     generateRouteSummary(day, request),
			DataSource:  "rule",
		}

		route.EstimatedCost = calculateRouteCost(route.Attractions)

		routes = append(routes, route)
	}

	return routes
}

func generateRouteTitle(day int, request model.TripRequest) string {
	if day == 1 {
		return "抵达与城市初体验"
	}

	if day == request.Days {
		return "收尾与返程准备"
	}

	return "核心景点游览"
}

func generateRouteSummary(day int, request model.TripRequest) string {
	if request.Preference == "轻松" {
		return "这一天安排相对轻松，避免过度奔波，适合慢节奏游玩。"
	}

	if request.Preference == "省钱" {
		return "这一天优先选择低成本景点和公共交通，控制整体预算。"
	}

	return "这一天根据目的地景点进行常规路线安排。"
}

func pickAttractionsForDay(day int, totalDays int, attractions []model.Attraction) []model.Attraction {
	if len(attractions) == 0 {
		return []model.Attraction{}
	}

	if totalDays <= 0 {
		totalDays = 1
	}

	if day <= 0 {
		day = 1
	}

	start := (day - 1) * len(attractions) / totalDays
	end := day * len(attractions) / totalDays

	if start >= len(attractions) {
		return []model.Attraction{}
	}

	if end > len(attractions) {
		end = len(attractions)
	}

	if start == end {
		end = start + 1
		if end > len(attractions) {
			end = len(attractions)
		}
	}

	return attractions[start:end]
}

func calculateRouteCost(attractions []model.Attraction) int {
	total := 0

	for _, attraction := range attractions {
		total += attraction.EstimatedCost
	}

	return total
}