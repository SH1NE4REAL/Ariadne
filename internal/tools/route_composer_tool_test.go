package tools

import (
	"testing"

	"ariadne/internal/model"
)

func TestComposeDailyRoutesKeepsFoodAsSupport(t *testing.T) {
	request := model.TripRequest{
		RawInput: "想吃本地小吃，轻松逛一逛",
		Days:     2,
	}
	profile := model.EffectivePreferenceProfile{
		Food: model.EffectiveDomainPreference{
			SoftPreferTags: []string{"food", "local_food"},
		},
		Route: model.EffectiveDomainPreference{
			SoftPreferTags: []string{"relaxed"},
		},
	}
	attractions := []model.Attraction{
		{
			Name:     "城市历史街区",
			Category: "景区",
			Address:  "市中心",
			Lat:      30.1,
			Lng:      120.1,
		},
		{
			Name:     "本地小吃街",
			Category: "美食",
			Address:  "市中心",
			Lat:      30.101,
			Lng:      120.101,
		},
		{
			Name:     "湖滨观景台",
			Category: "景点",
			Address:  "湖滨",
			Lat:      30.2,
			Lng:      120.2,
		},
	}

	routes := ComposeDailyRoutes(request, attractions, profile)

	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}

	for _, route := range routes {
		mainCount := 0
		foodCount := 0

		for _, attraction := range route.Attractions {
			poi := BuildTripPOI(attraction, 0)
			switch poi.Role {
			case tripPOIRoleMainAttraction:
				mainCount++
			case tripPOIRoleFoodSpot:
				foodCount++
			}
		}

		if foodCount > 0 && mainCount == 0 {
			t.Fatalf("route day %d has food spots but no main attraction: %#v", route.Day, route.Attractions)
		}
	}
}

func TestBuildTripPOIClassifiesMallAsShoppingSpot(t *testing.T) {
	poi := BuildTripPOI(model.Attraction{
		Name:     "普通购物中心",
		Category: "购物中心",
		Address:  "商业区",
	}, 0)

	if poi.Role != tripPOIRoleShoppingSpot {
		t.Fatalf("expected shopping spot, got %s with tags %#v", poi.Role, poi.Tags)
	}
}

func TestBuildTripPOIClassifiesRestaurantAsFoodSpot(t *testing.T) {
	poi := BuildTripPOI(model.Attraction{
		Name:     "普通饭店",
		Category: "餐厅",
		Address:  "市区",
	}, 0)

	if poi.Role != tripPOIRoleFoodSpot {
		t.Fatalf("expected food spot, got %s with tags %#v", poi.Role, poi.Tags)
	}
}

func TestBuildTripPOIDoesNotTreatPetMarketAsAquarium(t *testing.T) {
	poi := BuildTripPOI(model.Attraction{
		Name:     "花鸟鱼虫市场水族馆",
		Category: "市场",
		Address:  "市区",
	}, 0)

	if hasTag(poi.Tags, "aquarium") || poi.Role == tripPOIRoleMainAttraction {
		t.Fatalf("expected pet market not to be aquarium/main attraction, got role %s tags %#v", poi.Role, poi.Tags)
	}
}

func TestBuildPOIProfileFiltersClearlyInvalidPOIs(t *testing.T) {
	invalidCases := []model.Attraction{
		{Name: "某景区公交站", Category: "交通设施"},
		{Name: "某某口腔门诊", Category: "医疗保健"},
		{Name: "村委会", Category: "政府机构"},
		{Name: "照明工程公司", Category: "生活服务"},
		{Name: "安置房小区", Category: "住宅区"},
		{Name: "农贸市场", Category: "市场"},
	}

	for _, attraction := range invalidCases {
		profile := BuildPOIProfile(attraction)
		if !profile.Invalid {
			t.Fatalf("expected invalid POI for %#v, got tags %#v", attraction, profile.Tags)
		}
	}
}

func TestComposeDailyRoutesPrioritizesSeaIntent(t *testing.T) {
	request := model.TripRequest{
		RawInput: "想看海",
		Days:     1,
	}
	profile := model.EffectivePreferenceProfile{
		Attraction: model.EffectiveDomainPreference{
			SoftPreferTags: []string{"sea", "beach", "waterfront", "coast"},
		},
	}
	routes := ComposeDailyRoutes(request, []model.Attraction{
		{Name: "城市历史街区", Category: "景区"},
		{Name: "海滨步道", Category: "景点"},
	}, profile)

	if len(routes) != 1 {
		t.Fatalf("expected one route, got %d", len(routes))
	}

	foundSea := false
	for _, attraction := range routes[0].Attractions {
		if hasAnyTag(BuildTripPOI(attraction, 0).Tags, []string{"sea", "beach", "waterfront", "coast"}) {
			foundSea = true
			break
		}
	}

	if !foundSea {
		t.Fatalf("expected route to contain sea POI, got %#v", routes[0].Attractions)
	}
}

func TestComposeDailyRoutesFiltersShoppingWhenHardAvoided(t *testing.T) {
	request := model.TripRequest{
		RawInput: "不要普通商场，想看海",
		Days:     1,
	}
	profile := model.EffectivePreferenceProfile{
		Attraction: model.EffectiveDomainPreference{
			SoftPreferTags: []string{"sea", "beach"},
			HardAvoidTags:  []string{"shopping", "commercial_area"},
		},
	}

	routes := ComposeDailyRoutes(request, []model.Attraction{
		{Name: "百联滨江购物中心", Category: "购物:购物中心"},
		{Name: "海滨步道", Category: "旅游景点:海滩"},
	}, profile)

	for _, route := range routes {
		for _, attraction := range route.Attractions {
			if BuildTripPOI(attraction, 0).Role == tripPOIRoleShoppingSpot {
				t.Fatalf("shopping spot should not enter route when hard avoided: %#v", route.Attractions)
			}
		}
	}
}

func TestComposeDailyRoutesDoesNotUseFoodAsMainFallback(t *testing.T) {
	request := model.TripRequest{
		RawInput: "想吃本地小吃",
		Days:     1,
	}
	profile := model.EffectivePreferenceProfile{
		Food: model.EffectiveDomainPreference{
			SoftPreferTags: []string{"food", "local_food"},
		},
	}

	routes := ComposeDailyRoutes(request, []model.Attraction{
		{Name: "老字号饭店", Category: "美食:中餐厅"},
		{Name: "本地小吃店", Category: "美食:小吃快餐"},
	}, profile)

	if len(routes) != 1 {
		t.Fatalf("expected one route, got %d", len(routes))
	}

	if len(routes[0].Attractions) != 0 {
		t.Fatalf("food spots should not replace a missing main attraction: %#v", routes[0].Attractions)
	}
}

func TestComposeDailyRoutesFiltersMountainObservationWhenHardAvoided(t *testing.T) {
	request := model.TripRequest{
		RawInput: "不要爬山，轻松逛",
		Days:     1,
	}
	profile := model.EffectivePreferenceProfile{
		Attraction: model.EffectiveDomainPreference{
			HardAvoidTags:  []string{"mountain", "hiking", "high_exertion"},
			SoftPreferTags: []string{"city_walk"},
		},
	}

	routes := ComposeDailyRoutes(request, []model.Attraction{
		{Name: "山顶观景台", Category: "旅游景点:风景名胜"},
		{Name: "城市历史街区", Category: "旅游景点:风景名胜"},
	}, profile)

	for _, route := range routes {
		for _, attraction := range route.Attractions {
			if attraction.Name == "山顶观景台" {
				t.Fatalf("mountain-like POI should not enter route when hiking is hard avoided: %#v", route.Attractions)
			}
		}
	}
}

func TestSameDayBusinessRouteUsesLightTemplate(t *testing.T) {
	request := model.TripRequest{
		RawInput: "当天往返出差，会议后有空可以轻松走走",
		Days:     1,
	}
	profile := model.EffectivePreferenceProfile{
		Route: model.EffectiveDomainPreference{
			SoftPreferTags: []string{"business_trip", "same_day_return"},
		},
	}

	routes := ComposeDailyRoutes(request, []model.Attraction{
		{Name: "城市科技馆", Category: "文化场馆:科技馆"},
		{Name: "普通购物中心", Category: "购物:购物中心"},
		{Name: "城市地标滨水步道", Category: "旅游景点:风景名胜"},
	}, profile)

	if len(routes) != 1 {
		t.Fatalf("expected one same-day route, got %d", len(routes))
	}

	if len(routes[0].Attractions) > 1 {
		t.Fatalf("same-day business route should contain at most one light stop: %#v", routes[0].Attractions)
	}

	for _, attraction := range routes[0].Attractions {
		if attraction.Name == "城市科技馆" || attraction.Name == "普通购物中心" {
			t.Fatalf("same-day business route should not proactively use museum/science/shopping POIs: %#v", routes[0].Attractions)
		}
	}
}
