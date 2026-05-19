package tools

import (
	"testing"

	"ariadne/internal/model"
)

func TestBuildAttractionSearchKeywordsAddsMainFallbackForFoodOnly(t *testing.T) {
	tool := NewPreferenceAdapterTool()
	request := model.TripRequest{
		RawInput: "想吃本地小吃",
	}
	constraints := []model.PreferenceConstraint{
		{
			Domain:     "food",
			PreferTags: []string{"food", "local_food", "snack_street"},
			Strength:   "soft",
			Priority:   90,
			Source:     "current_request",
		},
	}

	keywords := tool.BuildAttractionSearchKeywords(request, constraints)

	if !containsKeyword(keywords, "景点") {
		t.Fatalf("expected generic main-attraction fallback keyword, got %#v", keywords)
	}

	if len(keywords) > 8 {
		t.Fatalf("expected keyword list to stay bounded, got %d: %#v", len(keywords), keywords)
	}
}

func TestBuildAttractionSearchKeywordsUsesGenericCityIndependentTerms(t *testing.T) {
	tool := NewPreferenceAdapterTool()
	request := model.TripRequest{
		RawInput: "想看海、逛老街、吃本地小吃、看夜景",
	}

	keywords := tool.BuildAttractionSearchKeywords(request, nil)

	for _, citySpecific := range []string{"锦里", "九眼桥", "太古里", "宽窄巷子", "春熙路"} {
		if containsKeyword(keywords, citySpecific) {
			t.Fatalf("did not expect city-specific keyword %q in %#v", citySpecific, keywords)
		}
	}

	for _, expected := range []string{"海边", "老街", "小吃街", "夜景"} {
		if !containsKeyword(keywords, expected) {
			t.Fatalf("expected keyword %q in %#v", expected, keywords)
		}
	}
}

func TestBuildFallbackAttractionSearchKeywordsForMissingSeaIntent(t *testing.T) {
	tool := NewPreferenceAdapterTool()
	request := model.TripRequest{
		RawInput: "想看海",
	}
	constraints := []model.PreferenceConstraint{
		{
			Domain:     "attraction",
			PreferTags: []string{"sea", "beach", "waterfront", "coast"},
			Strength:   "soft",
			Priority:   95,
			Source:     "current_request",
		},
	}

	keywords := tool.BuildFallbackAttractionSearchKeywords(request, []model.Attraction{
		{Name: "本地小吃街", Category: "美食"},
	}, constraints)

	if !containsKeyword(keywords, "海边") || !containsKeyword(keywords, "观海") {
		t.Fatalf("expected sea fallback keywords, got %#v", keywords)
	}
}

func containsKeyword(keywords []string, target string) bool {
	for _, keyword := range keywords {
		if keyword == target {
			return true
		}
	}

	return false
}
