package tools

import (
	"testing"

	"ariadne/internal/model"
)

func TestBuildHotelSearchTermsUsesUniqueStayPreference(t *testing.T) {
	terms := buildHotelSearchTerms(model.TripRequest{}, nil, []model.PreferenceConstraint{
		{
			Domain:     "hotel",
			PreferTags: []string{"homestay", "guesthouse", "unique_stay"},
			Strength:   "soft",
			Priority:   90,
			Source:     "current_request",
		},
	})

	for _, expected := range []string{"民宿", "客栈", "特色住宿"} {
		if !containsKeyword(terms, expected) {
			t.Fatalf("expected hotel search term %q in %#v", expected, terms)
		}
	}
}

func TestBuildHotelSearchTermsUsesHighEndPreference(t *testing.T) {
	terms := buildHotelSearchTerms(model.TripRequest{}, nil, []model.PreferenceConstraint{
		{
			Domain:     "hotel",
			PreferTags: []string{"high_end_hotel", "resort", "sea_nearby"},
			Strength:   "soft",
			Priority:   95,
			Source:     "current_request",
		},
	})

	for _, expected := range []string{"高端酒店", "度假酒店", "海景酒店", "海边酒店"} {
		if !containsKeyword(terms, expected) {
			t.Fatalf("expected hotel search term %q in %#v", expected, terms)
		}
	}
}
