package agent

import (
	"testing"

	"ariadne/internal/model"
)

func TestSelectRecommendedTrainOfferRequiresHighSpeedForCurrentPreference(t *testing.T) {
	request := model.TripRequest{
		RawInput:            "不要飞机，优先高铁或动车",
		TransportPreference: "高铁或动车",
	}
	constraints := []model.PreferenceConstraint{
		{
			Domain:     "transport",
			PreferTags: []string{"high_speed_train", "bullet_train", "train"},
			Strength:   "hard",
			Priority:   100,
			Source:     "current_request",
		},
	}

	offer := selectRecommendedTrainOffer(request, []model.TrainOffer{
		{
			Provider:             "fliggy",
			Status:               "ok",
			Price:                100,
			TotalDurationMinutes: 300,
			Segments: []model.TrainSegment{
				{TrainNo: "K123", TrainType: "普快", SeatClassName: "硬座"},
			},
		},
		{
			Provider:             "fliggy",
			Status:               "ok",
			Price:                300,
			TotalDurationMinutes: 160,
			Segments: []model.TrainSegment{
				{TrainNo: "G123", TrainType: "高铁", SeatClassName: "二等座"},
			},
		},
	}, "outbound", constraints)

	if offer.Status != "ok" || offer.Segments[0].TrainNo != "G123" {
		t.Fatalf("expected high-speed train recommendation, got %#v", offer)
	}
}

func TestSelectRecommendedTrainOfferDoesNotFallbackToSlowTrainForHighSpeedPreference(t *testing.T) {
	request := model.TripRequest{
		RawInput:            "只坐高铁",
		TransportPreference: "高铁",
	}

	offer := selectRecommendedTrainOffer(request, []model.TrainOffer{
		{
			Provider:             "fliggy",
			Status:               "ok",
			Price:                100,
			TotalDurationMinutes: 300,
			Segments: []model.TrainSegment{
				{TrainNo: "K123", TrainType: "普快", SeatClassName: "硬座"},
			},
		},
	}, "outbound", nil)

	if offer.Status != "unavailable" {
		t.Fatalf("expected unavailable instead of slow-train fallback, got %#v", offer)
	}
}

func TestSelectRecommendedTrainOfferPenalizesEarlyDeparture(t *testing.T) {
	request := model.TripRequest{RawInput: "轻松出行，优先高铁"}

	offer := selectRecommendedTrainOffer(request, []model.TrainOffer{
		{
			Provider:             "fliggy",
			Status:               "ok",
			Price:                260,
			TotalDurationMinutes: 120,
			Segments: []model.TrainSegment{
				{TrainNo: "G101", TrainType: "高铁", SeatClassName: "二等座", DepDateTime: "2026-06-01 05:30"},
			},
		},
		{
			Provider:             "fliggy",
			Status:               "ok",
			Price:                320,
			TotalDurationMinutes: 150,
			Segments: []model.TrainSegment{
				{TrainNo: "G102", TrainType: "高铁", SeatClassName: "二等座", DepDateTime: "2026-06-01 09:30"},
			},
		},
	}, "outbound", nil)

	if offer.Segments[0].TrainNo != "G102" {
		t.Fatalf("expected non-early train to win despite slightly higher price/duration, got %#v", offer)
	}
}
