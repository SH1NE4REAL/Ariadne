package tools

import (
	"strings"

	"ariadne/internal/model"
)

type RouteFeasibilityTool struct {
	Name        string
	Description string
}

func NewRouteFeasibilityTool() RouteFeasibilityTool {
	return RouteFeasibilityTool{
		Name:        "route_feasibility_tool",
		Description: "根据真实路段距离、通勤时间和用户偏好过滤不可行的每日路线",
	}
}

type feasibilityThreshold struct {
	MaxAttractionsPerDay int
	MaxTotalMinutes     int
	MaxSegmentMinutes   int
	MaxRemoteSegments   int
}

func (t RouteFeasibilityTool) Run(
	request model.TripRequest,
	dailyRoutes []model.DailyRoute,
) []model.DailyRoute {
	threshold := buildFeasibilityThreshold(request)

	result := make([]model.DailyRoute, 0, len(dailyRoutes))

	for _, route := range dailyRoutes {
		result = append(result, filterSingleDayRoute(route, threshold))
	}

	return result
}

func buildFeasibilityThreshold(request model.TripRequest) feasibilityThreshold {
	raw := request.RawInput

	if request.Preference == "轻松" ||
		strings.Contains(raw, "轻松") ||
		strings.Contains(raw, "不想太赶") ||
		strings.Contains(raw, "慢") {
		return feasibilityThreshold{
			MaxAttractionsPerDay: 2,
			MaxTotalMinutes:     90,
			MaxSegmentMinutes:   45,
			MaxRemoteSegments:   1,
		}
	}

	if request.Preference == "特种兵" ||
		strings.Contains(raw, "特种兵") ||
		strings.Contains(raw, "尽量多") ||
		strings.Contains(raw, "多玩") {
		return feasibilityThreshold{
			MaxAttractionsPerDay: 5,
			MaxTotalMinutes:     240,
			MaxSegmentMinutes:   120,
			MaxRemoteSegments:   2,
		}
	}

	return feasibilityThreshold{
		MaxAttractionsPerDay: 3,
		MaxTotalMinutes:     120,
		MaxSegmentMinutes:   60,
		MaxRemoteSegments:   1,
	}
}

func filterSingleDayRoute(
	route model.DailyRoute,
	threshold feasibilityThreshold,
) model.DailyRoute {
	if len(route.Attractions) <= 1 {
		return route
	}

	filteredAttractions := make([]model.Attraction, 0)
	filteredSegments := make([]model.RouteSegment, 0)

	filteredAttractions = append(filteredAttractions, route.Attractions[0])

	totalMinutes := 0
	remoteSegments := 0

	for i, segment := range route.RouteSegments {
		nextAttractionIndex := i + 1
		if nextAttractionIndex >= len(route.Attractions) {
			break
		}

		if len(filteredAttractions) >= threshold.MaxAttractionsPerDay {
			break
		}

		if segment.Status != "ok" {
			break
		}

		segmentMinutes := segment.DurationMinutes
		if segmentMinutes <= 0 && segment.DurationSeconds > 0 {
			segmentMinutes = segment.DurationSeconds / 60
		}

		isRemote := isRemoteSegment(segment)

		if isRemote && remoteSegments >= threshold.MaxRemoteSegments {
			break
		}

		if segmentMinutes > threshold.MaxSegmentMinutes && len(filteredAttractions) >= 1 {
			break
		}

		if totalMinutes+segmentMinutes > threshold.MaxTotalMinutes {
			break
		}

		filteredSegments = append(filteredSegments, segment)
		filteredAttractions = append(filteredAttractions, route.Attractions[nextAttractionIndex])

		totalMinutes += segmentMinutes
		if isRemote {
			remoteSegments++
		}
	}

	if len(filteredAttractions) == 0 {
		filteredAttractions = route.Attractions[:1]
	}

	route.Attractions = filteredAttractions
	route.RouteSegments = filteredSegments
	route.DataSource = route.DataSource + "_feasibility_filtered"
	route.Optimized = true

	route.Summary = buildFeasibilitySummary(route, totalMinutes)

	return route
}

func isRemoteSegment(segment model.RouteSegment) bool {
	if segment.DurationMinutes >= 60 {
		return true
	}

	if segment.DistanceMeters >= 35000 {
		return true
	}

	return false
}

func buildFeasibilitySummary(route model.DailyRoute, totalMinutes int) string {
	if len(route.Attractions) <= 1 {
		return "这一天保留一个核心景点，避免因景点距离过远导致行程过赶。"
	}

	if totalMinutes > 0 {
		return "这一天已根据真实通勤时间过滤过远景点，控制单日路线强度。"
	}

	return route.Summary
}