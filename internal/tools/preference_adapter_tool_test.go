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

	if !containsKeyword(keywords, "海边") || !containsKeyword(keywords, "海滩") {
		t.Fatalf("expected sea fallback keywords, got %#v", keywords)
	}
}

func TestBuildFallbackAttractionSearchKeywordsCoversCoreIntents(t *testing.T) {
	tool := NewPreferenceAdapterTool()
	constraints := []model.PreferenceConstraint{
		{
			Domain:     "food",
			PreferTags: []string{"food", "local_food", "snack_street"},
			Strength:   "soft",
			Priority:   90,
			Source:     "current_request",
		},
		{
			Domain:     "attraction",
			PreferTags: []string{"old_street", "night_view", "science_museum", "family"},
			Strength:   "soft",
			Priority:   90,
			Source:     "current_request",
		},
	}

	keywords := tool.BuildFallbackAttractionSearchKeywords(model.TripRequest{}, nil, constraints)

	for _, expected := range []string{"小吃街", "老街", "夜景", "科技馆"} {
		if !containsKeyword(keywords, expected) {
			t.Fatalf("expected fallback keyword %q in %#v", expected, keywords)
		}
	}

	if len(keywords) > 8 {
		t.Fatalf("expected fallback keyword list to stay bounded, got %#v", keywords)
	}
}

func TestBuildFallbackAttractionSearchGroupsSplitsByIntent(t *testing.T) {
	tool := NewPreferenceAdapterTool()
	constraints := []model.PreferenceConstraint{
		{
			Domain:     "food",
			PreferTags: []string{"food", "local_food", "snack_street"},
			Strength:   "soft",
			Priority:   90,
			Source:     "current_request",
		},
		{
			Domain:     "attraction",
			PreferTags: []string{"sea", "old_street", "night_view", "science_museum", "family"},
			Strength:   "soft",
			Priority:   95,
			Source:     "current_request",
		},
	}

	groups := tool.BuildFallbackAttractionSearchGroups(model.TripRequest{}, nil, constraints)
	intents := map[string][]string{}
	for _, group := range groups {
		intents[group.Intent] = group.Keywords
	}

	for _, intent := range []string{"sea", "local_food", "old_street", "night_view", "family_science"} {
		if len(intents[intent]) == 0 {
			t.Fatalf("expected fallback group for %s, got %#v", intent, groups)
		}
	}

	if !containsKeyword(intents["family_science"], "天文馆") ||
		!containsKeyword(intents["family_science"], "自然博物馆") {
		t.Fatalf("expected family science group to contain specific science keywords, got %#v", intents["family_science"])
	}
}

func TestBuildFallbackAttractionSearchKeywordsRemovesHardAvoidedShopping(t *testing.T) {
	tool := NewPreferenceAdapterTool()
	constraints := []model.PreferenceConstraint{
		{
			Domain:    "attraction",
			AvoidTags: []string{"shopping", "commercial_area"},
			Strength:  "hard",
			Priority:  100,
			Source:    "current_request",
		},
		{
			Domain:     "attraction",
			PreferTags: []string{"old_street", "city_walk"},
			Strength:   "soft",
			Priority:   90,
			Source:     "current_request",
		},
	}

	keywords := tool.BuildFallbackAttractionSearchKeywords(model.TripRequest{}, nil, constraints)

	for _, forbidden := range []string{"商场", "购物中心", "商业街", "市场"} {
		if containsKeyword(keywords, forbidden) {
			t.Fatalf("did not expect shopping keyword %q in %#v", forbidden, keywords)
		}
	}
}

func TestBuildAttractionSearchKeywordsCurrentRequestBeatsLongTermMuseumMemory(t *testing.T) {
	tool := NewPreferenceAdapterTool()
	request := model.TripRequest{
		RawInput: "想吃本地小吃、看夜景、逛老街",
	}
	constraints := []model.PreferenceConstraint{
		{
			Domain:         "food",
			PreferTags:     []string{"food", "local_food", "snack_street"},
			PreferKeywords: []string{"小吃街", "美食街"},
			Strength:       "soft",
			Priority:       90,
			Source:         "current_request",
		},
		{
			Domain:         "attraction",
			PreferTags:     []string{"museum", "culture"},
			PreferKeywords: []string{"博物馆", "展览馆", "历史", "人文"},
			Strength:       "soft",
			Priority:       60,
			Source:         "long_term_memory",
		},
	}

	keywords := tool.BuildAttractionSearchKeywords(request, constraints)

	for _, forbidden := range []string{"博物馆", "展览馆", "美术馆", "人文"} {
		if containsKeyword(keywords, forbidden) {
			t.Fatalf("did not expect long-term museum keyword %q in %#v", forbidden, keywords)
		}
	}

	for _, expected := range []string{"小吃街", "夜景", "老街"} {
		if !containsKeyword(keywords, expected) {
			t.Fatalf("expected current-request keyword %q in %#v", expected, keywords)
		}
	}
}

func TestBuildAttractionSearchKeywordsAllowsLongTermOnlyWhenCurrentRequestHasNoPOIIntent(t *testing.T) {
	tool := NewPreferenceAdapterTool()
	request := model.TripRequest{
		RawInput: "帮我安排周末行程",
	}
	constraints := []model.PreferenceConstraint{
		{
			Domain:         "attraction",
			PreferTags:     []string{"museum"},
			PreferKeywords: []string{"博物馆"},
			Strength:       "soft",
			Priority:       60,
			Source:         "long_term_memory",
		},
	}

	keywords := tool.BuildAttractionSearchKeywords(request, constraints)

	if !containsKeyword(keywords, "博物馆") {
		t.Fatalf("expected long-term keyword when current request has no POI intent, got %#v", keywords)
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
