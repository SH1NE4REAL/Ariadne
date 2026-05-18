package tools

import (
	"strings"
	"time"

	"ariadne/internal/model"
)

type TravelTimeWindowTool struct {
	Name        string
	Description string
}

func NewTravelTimeWindowTool() TravelTimeWindowTool {
	return TravelTimeWindowTool{
		Name:        "travel_time_window_tool",
		Description: "根据推荐交通方案的到达时间和返程时间调整首日、末日行程强度",
	}
}

func (t TravelTimeWindowTool) Run(
	request model.TripRequest,
	dailyRoutes []model.DailyRoute,
	recommendation model.TripRecommendation,
) []model.DailyRoute {
	if len(dailyRoutes) == 0 {
		return dailyRoutes
	}

	result := make([]model.DailyRoute, len(dailyRoutes))
	copy(result, dailyRoutes)

	arrivalTime, hasArrival := extractArrivalTime(recommendation)
	returnDepartureTime, hasReturnDeparture := extractReturnDepartureTime(recommendation)

	if hasArrival {
		result[0] = adjustFirstDayByArrivalTime(result[0], arrivalTime)
	}

	if hasReturnDeparture && len(result) > 1 {
		lastIndex := len(result) - 1
		result[lastIndex] = adjustLastDayByReturnTime(result[lastIndex], returnDepartureTime)
	}

	return result
}

func extractArrivalTime(recommendation model.TripRecommendation) (time.Time, bool) {
	if recommendation.RecommendedTransportType == "flight" {
		return extractFlightArrivalTime(recommendation.RecommendedFlight)
	}

	if recommendation.RecommendedTransportType == "train" {
		return extractTrainArrivalTime(recommendation.RecommendedOutboundTrain)
	}

	return time.Time{}, false
}

func extractReturnDepartureTime(recommendation model.TripRecommendation) (time.Time, bool) {
	if recommendation.RecommendedTransportType == "flight" {
		return extractFlightReturnDepartureTime(recommendation.RecommendedFlight)
	}

	if recommendation.RecommendedTransportType == "train" {
		return extractTrainReturnDepartureTime(recommendation.RecommendedReturnTrain)
	}

	return time.Time{}, false
}

func extractFlightArrivalTime(offer model.FlightOffer) (time.Time, bool) {
	for _, journey := range offer.Journeys {
		if journey.Direction != "outbound" {
			continue
		}

		if len(journey.Segments) == 0 {
			return time.Time{}, false
		}

		lastSegment := journey.Segments[len(journey.Segments)-1]
		return parseDateTime(lastSegment.ArrDateTime)
	}

	return time.Time{}, false
}

func extractFlightReturnDepartureTime(offer model.FlightOffer) (time.Time, bool) {
	for _, journey := range offer.Journeys {
		if journey.Direction != "return" {
			continue
		}

		if len(journey.Segments) == 0 {
			return time.Time{}, false
		}

		firstSegment := journey.Segments[0]
		return parseDateTime(firstSegment.DepDateTime)
	}

	return time.Time{}, false
}

func extractTrainArrivalTime(offer model.TrainOffer) (time.Time, bool) {
	if len(offer.Segments) == 0 {
		return time.Time{}, false
	}

	lastSegment := offer.Segments[len(offer.Segments)-1]
	return parseDateTime(lastSegment.ArrDateTime)
}

func extractTrainReturnDepartureTime(offer model.TrainOffer) (time.Time, bool) {
	if len(offer.Segments) == 0 {
		return time.Time{}, false
	}

	firstSegment := offer.Segments[0]
	return parseDateTime(firstSegment.DepDateTime)
}

func parseDateTime(text string) (time.Time, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return time.Time{}, false
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, text, time.Local)
		if err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func adjustFirstDayByArrivalTime(route model.DailyRoute, arrivalTime time.Time) model.DailyRoute {
	hour := arrivalTime.Hour()

	maxAttractions := 2
	reason := "这一天已根据抵达时间适当压缩行程。"

	if hour < 6 {
		maxAttractions = 1
		reason = "由于推荐交通方案为凌晨抵达，首日只保留一个轻量核心景点，避免过度疲劳。"
	} else if hour < 12 {
		maxAttractions = 2
		reason = "由于上午抵达，首日保留少量景点，避免刚到目的地就过度奔波。"
	} else if hour < 16 {
		maxAttractions = 1
		reason = "由于下午抵达，首日只安排一个轻量景点。"
	} else {
		maxAttractions = 1
		reason = "由于傍晚或夜间抵达，首日主要建议办理入住和休息，仅保留一个轻量景点。"
	}

	route = keepFirstNAttractions(route, maxAttractions)
	route.Summary = reason
	route.DataSource = appendDataSourceSuffix(route.DataSource, "time_window_filtered")

	return route
}

func adjustLastDayByReturnTime(route model.DailyRoute, returnDepartureTime time.Time) model.DailyRoute {
	hour := returnDepartureTime.Hour()

	maxAttractions := 2
	reason := "这一天已根据返程时间控制行程强度。"

	if hour < 10 {
		maxAttractions = 0
		reason = "由于返程时间较早，末日不安排景点，建议直接前往机场或车站。"
	} else if hour < 14 {
		maxAttractions = 1
		reason = "由于中午前后返程，末日只保留一个轻量景点。"
	} else if hour < 18 {
		maxAttractions = 1
		reason = "由于下午返程，末日只安排一个核心景点，并预留前往机场或车站的时间。"
	} else {
		maxAttractions = 2
		reason = "由于晚上返程，末日可保留一到两个景点，但仍需预留前往机场或车站的时间。"
	}

	route = keepFirstNAttractions(route, maxAttractions)
	route.Summary = reason
	route.DataSource = appendDataSourceSuffix(route.DataSource, "time_window_filtered")

	return route
}

func keepFirstNAttractions(route model.DailyRoute, n int) model.DailyRoute {
	if n < 0 {
		n = 0
	}

	if n >= len(route.Attractions) {
		return route
	}

	route.Attractions = route.Attractions[:n]

	if n <= 1 {
		route.RouteSegments = []model.RouteSegment{}
		return route
	}

	maxSegments := n - 1
	if maxSegments < len(route.RouteSegments) {
		route.RouteSegments = route.RouteSegments[:maxSegments]
	}

	return route
}

func appendDataSourceSuffix(dataSource string, suffix string) string {
	if dataSource == "" {
		return suffix
	}

	if strings.Contains(dataSource, suffix) {
		return dataSource
	}

	return dataSource + "_" + suffix
}