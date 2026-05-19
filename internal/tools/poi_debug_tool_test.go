package tools

import (
	"testing"

	"ariadne/internal/model"
)

func TestBuildPOIDebugReportCountsRejectedReasons(t *testing.T) {
	raw := []model.Attraction{
		{Name: "城市科技馆", Category: "文化场馆:科技馆"},
		{Name: "普通购物中心", Category: "购物:购物中心"},
		{Name: "景区公交站", Category: "交通设施"},
	}
	final := []model.Attraction{
		{Name: "城市科技馆", Category: "文化场馆:科技馆"},
	}
	profile := model.EffectivePreferenceProfile{
		Attraction: model.EffectiveDomainPreference{
			HardAvoidTags: []string{"shopping", "commercial_area"},
		},
	}

	report := BuildPOIDebugReport(
		raw,
		final,
		[]string{"科技馆"},
		map[string]int{"科技馆": 3},
		map[string]string{"科技馆": "api_ok", "海洋馆": "api_empty"},
		map[string]string{"夜景": "timeout"},
		profile,
	)

	if report.RawPOICount != 3 {
		t.Fatalf("expected raw count 3, got %#v", report)
	}

	if report.FinalRoutablePOICount != 1 {
		t.Fatalf("expected one final routable POI, got %#v", report)
	}

	if report.RejectedReasons["hard_avoid_shopping"] == 0 {
		t.Fatalf("expected hard_avoid_shopping rejection, got %#v", report.RejectedReasons)
	}

	if report.RejectedReasons["invalid_poi"] == 0 {
		t.Fatalf("expected invalid_poi rejection, got %#v", report.RejectedReasons)
	}

	if report.RawPOICountByKeyword["科技馆"] != 3 {
		t.Fatalf("expected raw count by keyword, got %#v", report.RawPOICountByKeyword)
	}

	if report.SearchStatusByKeyword["海洋馆"] != "api_empty" {
		t.Fatalf("expected api_empty status to be preserved, got %#v", report.SearchStatusByKeyword)
	}

	if report.SearchErrorByKeyword["夜景"] == "" {
		t.Fatalf("expected search error to be preserved, got %#v", report.SearchErrorByKeyword)
	}
}
