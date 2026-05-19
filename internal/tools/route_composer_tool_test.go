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
