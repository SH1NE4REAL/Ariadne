package tools

import (
	"strings"

	"ariadne/internal/model"
)

func EvaluateFeasibility(
	request model.TripRequest,
	routeDistance model.RouteDistance,
	recommendation model.TripRecommendation,
	profile model.EffectivePreferenceProfile,
) []model.FeasibilityIssue {
	issues := make([]model.FeasibilityIssue, 0)

	if routeDistance.DistanceMeters > 1500000 &&
		hasString(profile.Transport.HardAvoidTags, "flight") &&
		request.Days <= 2 {
		issues = append(issues, model.FeasibilityIssue{
			Level:   "severe",
			Code:    "long_distance_no_flight_short_trip",
			Message: "Long-distance trip with flight forbidden and short duration may be unrealistic.",
		})
	}

	if request.Budget > 0 && request.Budget < 1000 && hotelPrefersHighEnd(profile.Hotel) {
		issues = append(issues, model.FeasibilityIssue{
			Level:   "impossible",
			Code:    "low_budget_high_end_hotel",
			Message: "Budget is likely incompatible with high-end or resort hotel requirements.",
		})
	}

	if recommendation.RecommendedTransportType == "train" &&
		(recommendation.RecommendedOutboundTrain.Status != "ok" ||
			(needsReturnTransport(request) && recommendation.RecommendedReturnTrain.Status != "ok")) {
		issues = append(issues, model.FeasibilityIssue{
			Level:   "severe",
			Code:    "recommended_train_unavailable",
			Message: "Recommended train transport is missing available real ticket data.",
		})
	}

	if tripNeedsHotel(request) && recommendation.RecommendedHotel.Status != "ok" {
		issues = append(issues, model.FeasibilityIssue{
			Level:   "warning",
			Code:    "hotel_unavailable",
			Message: "No available hotel satisfies the structured recommendation.",
		})
	}

	if recommendation.TotalRealCost == 0 && criticalRecommendationDataUnavailable(request, recommendation) {
		issues = append(issues, model.FeasibilityIssue{
			Level:   "warning",
			Code:    "zero_cost_due_to_unavailable_data",
			Message: "Total real cost is zero because key transport or hotel data is unavailable.",
		})
	}

	return issues
}

func hotelPrefersHighEnd(preference model.EffectiveDomainPreference) bool {
	for _, tag := range []string{"high_end_hotel", "resort", "comfort_hotel"} {
		if hasString(preference.HardPreferTags, tag) || hasString(preference.SoftPreferTags, tag) {
			return true
		}
	}

	return false
}

func tripNeedsHotel(request model.TripRequest) bool {
	if request.Days < 2 {
		return false
	}

	text := strings.ToLower(request.RawInput + " " + request.Preference)
	return !containsAnyText(text, []string{"当天往返", "不住宿", "不需要酒店", "不用酒店", "不住酒店"})
}

func needsReturnTransport(request model.TripRequest) bool {
	return request.Days > 1
}

func criticalRecommendationDataUnavailable(request model.TripRequest, recommendation model.TripRecommendation) bool {
	if recommendation.RecommendedTransportType == "train" {
		if recommendation.RecommendedOutboundTrain.Status != "ok" {
			return true
		}
		if needsReturnTransport(request) && recommendation.RecommendedReturnTrain.Status != "ok" {
			return true
		}
	}

	if recommendation.RecommendedTransportType == "flight" && recommendation.RecommendedFlight.Status != "ok" {
		return true
	}

	return tripNeedsHotel(request) && recommendation.RecommendedHotel.Status != "ok"
}
