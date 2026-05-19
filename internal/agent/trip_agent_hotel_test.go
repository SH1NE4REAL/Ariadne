package agent

import (
	"testing"

	"ariadne/internal/model"
)

func TestSelectRecommendedHotelHonorsCurrentHighEndPreference(t *testing.T) {
	request := model.TripRequest{RawInput: "想住好一点，优先海边高端酒店或度假酒店", Days: 2}
	constraints := []model.PreferenceConstraint{
		{
			Domain:     "hotel",
			PreferTags: []string{"high_end_hotel", "resort", "sea_nearby", "comfort_hotel"},
			Strength:   "soft",
			Priority:   100,
			Source:     "current_request",
		},
	}

	hotel := selectRecommendedHotel(request, model.BudgetBreakdown{}, []model.HotelOffer{
		{
			Name:          "经济型连锁酒店",
			Star:          "经济型",
			TotalPrice:    300,
			PricePerNight: 150,
			Status:        "ok",
		},
		{
			Name:          "海边度假酒店",
			Star:          "五星",
			NearbyPOI:     "海滨",
			TotalPrice:    1600,
			PricePerNight: 800,
			Status:        "ok",
		},
	}, constraints, nil)

	if hotel.Name != "海边度假酒店" {
		t.Fatalf("expected high-end sea resort, got %#v", hotel)
	}
}

func TestSelectRecommendedHotelReturnsUnavailableWhenCurrentPreferenceUnmatched(t *testing.T) {
	request := model.TripRequest{RawInput: "想体验有特色的海边民宿或客栈", Days: 2}
	constraints := []model.PreferenceConstraint{
		{
			Domain:     "hotel",
			PreferTags: []string{"homestay", "guesthouse", "sea_nearby", "unique_stay"},
			Strength:   "soft",
			Priority:   100,
			Source:     "current_request",
		},
	}

	hotel := selectRecommendedHotel(request, model.BudgetBreakdown{}, []model.HotelOffer{
		{
			Name:          "普通连锁酒店",
			Star:          "经济型",
			TotalPrice:    300,
			PricePerNight: 150,
			Status:        "ok",
		},
	}, constraints, nil)

	if hotel.Status != "unavailable" {
		t.Fatalf("expected unavailable when no hotel matches current lodging preference, got %#v", hotel)
	}
}

func TestSelectRecommendedHotelFiltersHardAvoidAccommodationTags(t *testing.T) {
	request := model.TripRequest{RawInput: "不要青旅、多人间、床位房", Days: 2}
	constraints := []model.PreferenceConstraint{
		{
			Domain:          "hotel",
			AvoidTags:       []string{"hostel", "apartment"},
			ExcludeKeywords: []string{"青旅", "多人间", "床位房", "公寓"},
			Strength:        "hard",
			Priority:        100,
			Source:          "current_request",
		},
	}

	hotel := selectRecommendedHotel(request, model.BudgetBreakdown{}, []model.HotelOffer{
		{
			Name:          "青年旅舍多人间",
			TotalPrice:    120,
			PricePerNight: 60,
			Status:        "ok",
		},
		{
			Name:          "标准酒店",
			TotalPrice:    500,
			PricePerNight: 250,
			Status:        "ok",
		},
	}, constraints, nil)

	if hotel.Name != "标准酒店" {
		t.Fatalf("expected hard-avoided hostel to be filtered, got %#v", hotel)
	}
}
