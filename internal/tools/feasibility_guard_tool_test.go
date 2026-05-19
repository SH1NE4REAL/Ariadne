package tools

import (
	"testing"

	"ariadne/internal/model"
)

func TestEvaluateFeasibilityLongDistanceNoFlightShortTrip(t *testing.T) {
	issues := EvaluateFeasibility(
		model.TripRequest{Days: 2},
		model.RouteDistance{DistanceMeters: 1800000, Status: "ok"},
		model.TripRecommendation{RecommendedTransportType: "train"},
		model.EffectivePreferenceProfile{
			Transport: model.EffectiveDomainPreference{
				HardAvoidTags: []string{"flight"},
			},
		},
	)

	if !hasFeasibilityCode(issues, "long_distance_no_flight_short_trip") {
		t.Fatalf("expected long-distance no-flight issue, got %#v", issues)
	}
}

func TestEvaluateFeasibilityLowBudgetHighEndHotel(t *testing.T) {
	issues := EvaluateFeasibility(
		model.TripRequest{Days: 2, Budget: 800},
		model.RouteDistance{},
		model.TripRecommendation{RecommendedTransportType: "train"},
		model.EffectivePreferenceProfile{
			Hotel: model.EffectiveDomainPreference{
				SoftPreferTags: []string{"high_end_hotel", "resort"},
			},
		},
	)

	if !hasFeasibilityCode(issues, "low_budget_high_end_hotel") {
		t.Fatalf("expected low-budget high-end hotel issue, got %#v", issues)
	}
}

func TestEvaluateFeasibilityZeroCostUnavailableData(t *testing.T) {
	issues := EvaluateFeasibility(
		model.TripRequest{Days: 2},
		model.RouteDistance{},
		model.TripRecommendation{
			RecommendedTransportType: "train",
			TotalRealCost:            0,
			RecommendedOutboundTrain: model.TrainOffer{Status: "unavailable"},
			RecommendedReturnTrain:   model.TrainOffer{Status: "unavailable"},
			RecommendedHotel:         model.HotelOffer{Status: "unavailable"},
		},
		model.EffectivePreferenceProfile{},
	)

	if !hasFeasibilityCode(issues, "zero_cost_due_to_unavailable_data") {
		t.Fatalf("expected zero-cost unavailable-data issue, got %#v", issues)
	}
}

func hasFeasibilityCode(issues []model.FeasibilityIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}

	return false
}
