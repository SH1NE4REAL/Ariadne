package tools

import "ariadne/internal/model"

type RouteOptimizerTool struct {
	Name        string
	Description string
}

func NewRouteOptimizerTool() RouteOptimizerTool {
	return RouteOptimizerTool{
		Name:        "route_optimizer_tool",
		Description: "根据腾讯距离矩阵结果优化每日景点游览顺序",
	}
}

func (t RouteOptimizerTool) Run(routes []model.DailyRoute, mapConfig model.MapConfig) []model.DailyRoute {
	for i := range routes {
		attractions := routes[i].Attractions

		if len(attractions) < 2 {
			routes[i].Optimized = false
			routes[i].OptimizationStrategy = "not_enough_attractions"
			continue
		}

		optimizedAttractions, routeSegments := optimizeAttractionsByNearestNeighbor(attractions, mapConfig)

		routes[i].Attractions = optimizedAttractions
		routes[i].RouteSegments = routeSegments
		routes[i].Optimized = true
		routes[i].OptimizationStrategy = "nearest_neighbor_by_tencent_distance"
		routes[i].DataSource = "tencent_map_optimized"
	}

	return routes
}

func optimizeAttractionsByNearestNeighbor(
	attractions []model.Attraction,
	mapConfig model.MapConfig,
) ([]model.Attraction, []model.RouteSegment) {
	if len(attractions) < 2 {
		return attractions, []model.RouteSegment{}
	}

	ordered := make([]model.Attraction, 0)
	segments := make([]model.RouteSegment, 0)

	remaining := make([]model.Attraction, 0)
	remaining = append(remaining, attractions...)

	current := remaining[0]
	ordered = append(ordered, current)
	remaining = remaining[1:]

	for len(remaining) > 0 {
		bestIndex := 0
		var bestSegment model.RouteSegment
		bestScore := 0
		foundAvailableSegment := false

		for i, candidate := range remaining {
			segment := queryRouteSegmentWithTencent(current, candidate, mapConfig)

			if segment.Status != "ok" {
				continue
			}

			score := segment.DurationSeconds
			if score <= 0 {
				score = segment.DistanceMeters
			}

			if !foundAvailableSegment || score < bestScore {
				foundAvailableSegment = true
				bestScore = score
				bestIndex = i
				bestSegment = segment
			}
		}

		next := remaining[bestIndex]

		if !foundAvailableSegment {
			bestSegment = queryRouteSegmentWithTencent(current, next, mapConfig)
		}

		segments = append(segments, bestSegment)
		ordered = append(ordered, next)

		current = next

		remaining = append(remaining[:bestIndex], remaining[bestIndex+1:]...)
	}

	return ordered, segments
}