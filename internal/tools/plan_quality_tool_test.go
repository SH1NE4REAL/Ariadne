package tools

import (
	"testing"

	"ariadne/internal/model"
)

func TestPlanQualityReportWarnsWhenCoreIntentMissing(t *testing.T) {
	report := BuildPlanQualityReport(
		[]model.DailyRoute{
			{
				Day: 1,
				Attractions: []model.Attraction{
					{Name: "本地小吃街", Category: "美食:小吃快餐"},
				},
			},
		},
		nil,
		model.TripRecommendation{
			RecommendedTransportType: "train",
			BudgetStatus:             "ok",
		},
		nil,
		nil,
		model.EffectivePreferenceProfile{
			Attraction: model.EffectiveDomainPreference{
				SoftPreferTags: []string{"sea", "beach"},
			},
		},
	)

	if report.Score >= 100 {
		t.Fatalf("expected quality score to drop, got %#v", report)
	}

	if report.CoreIntentCoverage["sea"] {
		t.Fatalf("expected sea intent to be uncovered, got %#v", report.CoreIntentCoverage)
	}

	if len(report.Warnings) == 0 {
		t.Fatalf("expected warnings for missing core intent")
	}
}

func TestPlanQualityReportWarnsForShoppingHardAvoid(t *testing.T) {
	report := BuildPlanQualityReport(
		[]model.DailyRoute{
			{
				Day: 1,
				Attractions: []model.Attraction{
					{Name: "普通购物中心", Category: "购物:购物中心"},
				},
			},
		},
		nil,
		model.TripRecommendation{
			RecommendedTransportType: "train",
			BudgetStatus:             "ok",
		},
		nil,
		nil,
		model.EffectivePreferenceProfile{
			Attraction: model.EffectiveDomainPreference{
				HardAvoidTags: []string{"shopping", "commercial_area"},
			},
		},
	)

	if report.HardConstraintPassed {
		t.Fatalf("expected hard constraint failure, got %#v", report)
	}

	if report.Score > 80 {
		t.Fatalf("expected shopping violation to lower score, got %#v", report)
	}
}

func TestPlanQualityFamilyScienceRequiresRealSciencePOI(t *testing.T) {
	report := BuildPlanQualityReport(
		[]model.DailyRoute{
			{
				Day: 1,
				Attractions: []model.Attraction{
					{Name: "城市历史博物馆", Category: "文化场馆:博物馆"},
					{Name: "普通美术馆", Category: "文化场馆:美术馆"},
				},
			},
		},
		nil,
		model.TripRecommendation{
			RecommendedTransportType: "train",
			BudgetStatus:             "ok",
		},
		nil,
		nil,
		model.EffectivePreferenceProfile{
			Attraction: model.EffectiveDomainPreference{
				SoftPreferTags: []string{"family", "indoor", "science_museum", "aquarium"},
			},
		},
	)

	if report.CoreIntentCoverage["family_science"] {
		t.Fatalf("history/art museums should not satisfy family_science coverage: %#v", report.CoreIntentCoverage)
	}

	if report.Score >= 100 {
		t.Fatalf("expected score to drop when family science intent is missing, got %#v", report)
	}
}
