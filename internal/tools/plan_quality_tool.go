package tools

import (
	"math"
	"strings"

	"ariadne/internal/model"
)

func BuildPlanQualityReport(
	dailyRoutes []model.DailyRoute,
	attractions []model.Attraction,
	recommendation model.TripRecommendation,
	violations []model.RecommendationViolation,
	feasibilityIssues []model.FeasibilityIssue,
	preferenceProfile model.EffectivePreferenceProfile,
) model.PlanQualityReport {
	report := model.PlanQualityReport{
		HardConstraintPassed:     true,
		SummaryConsistencyPassed: true,
		BudgetFeasibility:        "unknown",
		CoreIntentCoverage:       map[string]bool{},
		FeasibilityIssues:        feasibilityIssues,
		Warnings:                 make([]string, 0),
		Score:                    100,
	}

	for _, violation := range violations {
		if strings.Contains(strings.ToLower(violation.Type), "hard_constraint") {
			report.HardConstraintPassed = false
			report.Warnings = append(report.Warnings, violation.Message)
			report.Score -= 30
		}
	}

	applyFeasibilityIssues(&report, feasibilityIssues)

	if hasString(preferenceProfile.Transport.HardAvoidTags, "flight") &&
		recommendation.RecommendedTransportType == "flight" {
		report.HardConstraintPassed = false
		report.Warnings = append(report.Warnings, "Flight recommendation violates user's hard transport preference.")
		report.Score -= 30
	}

	for _, attraction := range attractions {
		poi := BuildTripPOI(attraction, 0)
		if poi.Role == tripPOIRoleInvalid {
			report.InvalidPOICount++
		}
	}

	totalTransferMinutes := 0
	transferCount := 0

	for _, route := range dailyRoutes {
		dayMainCount := 0
		dayFoodCount := 0

		for _, attraction := range route.Attractions {
			poi := BuildTripPOI(attraction, 0)
			switch poi.Role {
			case tripPOIRoleMainAttraction:
				report.MainAttractionCount++
				dayMainCount++
			case tripPOIRoleFoodSpot:
				report.FoodSpotCount++
				dayFoodCount++
			case tripPOIRoleShoppingSpot:
				report.ShoppingSpotCount++
			case tripPOIRoleInvalid:
				report.InvalidPOICount++
				report.Warnings = append(report.Warnings, "Invalid POI entered daily routes.")
				report.Score -= 20
			}
		}

		if dayMainCount == 0 && dayFoodCount > 0 {
			report.Warnings = append(report.Warnings, "A daily route only contains food spots and lacks a main attraction.")
			report.Score -= 12
		}

		for _, segment := range route.RouteSegments {
			minutes := segment.DurationMinutes
			if minutes <= 0 && segment.DurationSeconds > 0 {
				minutes = segment.DurationSeconds / 60
			}
			if minutes <= 0 {
				continue
			}

			totalTransferMinutes += minutes
			transferCount++
		}
	}

	if transferCount > 0 {
		report.AvgTransferMinutes = totalTransferMinutes / transferCount
	}

	if len(dailyRoutes) > 0 && report.MainAttractionCount < len(dailyRoutes) {
		report.Warnings = append(report.Warnings, "Some days do not have a main attraction.")
		report.Score -= 10
	}

	if report.MainAttractionCount == 0 {
		report.Warnings = append(report.Warnings, "The route lacks main attractions and currently looks more like a food-only plan.")
		report.Score -= 25
	}

	applyCoreIntentCoverage(&report, dailyRoutes, preferenceProfile)

	if report.FoodSpotCount > report.MainAttractionCount && report.MainAttractionCount > 0 {
		report.Warnings = append(report.Warnings, "Food spots outnumber main attractions; route composition may be too food-heavy.")
		report.Score -= 8
	}

	if hasString(preferenceProfile.Attraction.HardAvoidTags, "shopping") ||
		hasString(preferenceProfile.Attraction.HardAvoidTags, "commercial_area") {
		if report.ShoppingSpotCount > 0 {
			report.HardConstraintPassed = false
			report.Warnings = append(report.Warnings, "Shopping POI violates user preference.")
			report.Score -= 20
		}
	}

	if report.InvalidPOICount > 0 {
		report.Warnings = append(report.Warnings, "Invalid or non-travel POIs were detected in the candidate set.")
		report.Score -= report.InvalidPOICount * 5
	}

	if report.AvgTransferMinutes > 60 {
		report.Warnings = append(report.Warnings, "Average transfer time is high.")
		report.Score -= 10
	}

	report.HotelDistanceToCoreKM = hotelDistanceToRouteCore(recommendation.RecommendedHotel, dailyRoutes)
	if report.HotelDistanceToCoreKM > 50 {
		report.Warnings = append(report.Warnings, "Recommended hotel is very far from the route core.")
		report.Score -= 20
	} else if report.HotelDistanceToCoreKM > 20 {
		report.Warnings = append(report.Warnings, "Recommended hotel is far from the route core.")
		report.Score -= 10
	} else if report.HotelDistanceToCoreKM > 15 {
		report.Warnings = append(report.Warnings, "Recommended hotel is far from the route core.")
		report.Score -= 15
	} else if report.HotelDistanceToCoreKM > 8 {
		report.Warnings = append(report.Warnings, "Recommended hotel is not close to the route core.")
		report.Score -= 8
	}

	applyHotelQualityChecks(&report, recommendation.RecommendedHotel, preferenceProfile)
	applyBudgetFeasibility(&report, recommendation)

	if report.Score < 0 {
		report.Score = 0
	}

	return report
}

func applyFeasibilityIssues(report *model.PlanQualityReport, issues []model.FeasibilityIssue) {
	for _, issue := range issues {
		if strings.TrimSpace(issue.Message) != "" {
			report.Warnings = append(report.Warnings, issue.Message)
		}

		switch strings.ToLower(issue.Level) {
		case "impossible":
			report.Score -= 30
			report.BudgetFeasibility = "impossible"
		case "severe":
			report.Score -= 20
		default:
			report.Score -= 10
		}
	}
}

func applyCoreIntentCoverage(
	report *model.PlanQualityReport,
	dailyRoutes []model.DailyRoute,
	preferenceProfile model.EffectivePreferenceProfile,
) {
	checks := []struct {
		name        string
		intentTags  []string
		routeTags   []string
		warning     string
		mainOnly    bool
		scoreImpact int
	}{
		{
			name:        "sea",
			intentTags:  []string{"sea", "beach", "waterfront", "coast"},
			routeTags:   []string{"sea", "beach", "waterfront", "coast"},
			warning:     "Core sea/beach intent is not covered.",
			mainOnly:    true,
			scoreImpact: 15,
		},
		{
			name:        "food",
			intentTags:  []string{"food", "local_food", "snack_street"},
			routeTags:   []string{"food", "local_food", "snack_street"},
			warning:     "Food intent is not covered.",
			mainOnly:    false,
			scoreImpact: 15,
		},
		{
			name:        "night_view",
			intentTags:  []string{"night_view", "landmark"},
			routeTags:   []string{"night_view", "landmark", "waterfront"},
			warning:     "Night view intent is not covered.",
			mainOnly:    false,
			scoreImpact: 15,
		},
		{
			name:        "family_science",
			intentTags:  []string{"family", "science_museum", "aquarium", "indoor", "child_friendly"},
			routeTags:   []string{"science_museum", "astronomy", "aquarium", "children_exhibition", "natural_science"},
			warning:     "Family/science intent is not covered.",
			mainOnly:    true,
			scoreImpact: 15,
		},
		{
			name:        "old_street",
			intentTags:  []string{"old_street", "city_walk", "historic_site"},
			routeTags:   []string{"old_street", "city_walk", "historic_site"},
			warning:     "Old street/city walk intent is not covered.",
			mainOnly:    true,
			scoreImpact: 15,
		},
	}

	for _, check := range checks {
		if !effectiveProfileContainsAnyTag(preferenceProfile, check.intentTags) {
			continue
		}

		covered := dailyRoutesContainAnyTagWithRole(dailyRoutes, check.routeTags, check.mainOnly)
		report.CoreIntentCoverage[check.name] = covered
		if !covered {
			report.Warnings = append(report.Warnings, check.warning)
			report.Score -= check.scoreImpact
		}
	}
}

func applyHotelQualityChecks(
	report *model.PlanQualityReport,
	hotel model.HotelOffer,
	preferenceProfile model.EffectivePreferenceProfile,
) {
	if hotel.Status != "ok" || strings.TrimSpace(hotel.Name) == "" {
		if len(preferenceProfile.Hotel.HardPreferTags) > 0 || len(preferenceProfile.Hotel.SoftPreferTags) > 0 {
			report.Warnings = append(report.Warnings, "No structured hotel recommendation is available.")
			report.Score -= 20
		}
		return
	}

	hotelTags := inferHotelTagsForQuality(hotel)
	for _, tag := range preferenceProfile.Hotel.HardAvoidTags {
		if hasString(hotelTags, tag) {
			report.HardConstraintPassed = false
			report.Warnings = append(report.Warnings, "Recommended hotel violates user's hard accommodation preference.")
			report.Score -= 30
			return
		}
	}
}

func applyBudgetFeasibility(report *model.PlanQualityReport, recommendation model.TripRecommendation) {
	switch recommendation.BudgetStatus {
	case "ok":
		report.BudgetFeasibility = "ok"
	case "over_budget":
		report.BudgetFeasibility = "tight"
		report.Warnings = append(report.Warnings, "Budget is tight or over the requested amount.")
		report.Score -= 10
	default:
		report.BudgetFeasibility = "unknown"
	}

	if recommendation.Budget > 0 && recommendation.TotalRealCost > 0 {
		if recommendation.TotalRealCost > recommendation.Budget*2 {
			report.BudgetFeasibility = "impossible"
			report.Warnings = append(report.Warnings, "Budget is not feasible under current hard constraints.")
			report.Score -= 30
		} else if recommendation.TotalRealCost > recommendation.Budget {
			report.BudgetFeasibility = "tight"
		}
	}
}

func effectiveProfileContainsAnyTag(profile model.EffectivePreferenceProfile, tags []string) bool {
	for _, tag := range tags {
		if hasString(profile.Attraction.HardPreferTags, tag) ||
			hasString(profile.Attraction.SoftPreferTags, tag) ||
			hasString(profile.Food.HardPreferTags, tag) ||
			hasString(profile.Food.SoftPreferTags, tag) ||
			hasString(profile.Route.HardPreferTags, tag) ||
			hasString(profile.Route.SoftPreferTags, tag) {
			return true
		}
	}

	return false
}

func dailyRoutesContainAnyTag(dailyRoutes []model.DailyRoute, tags []string) bool {
	return dailyRoutesContainAnyTagWithRole(dailyRoutes, tags, false)
}

func dailyRoutesContainAnyTagWithRole(dailyRoutes []model.DailyRoute, tags []string, mainOnly bool) bool {
	for _, route := range dailyRoutes {
		for _, attraction := range route.Attractions {
			poi := BuildTripPOI(attraction, 0)
			if mainOnly && poi.Role != tripPOIRoleMainAttraction {
				continue
			}

			if hasAnyTag(poi.Tags, tags) {
				return true
			}
		}
	}

	return false
}

func inferHotelTagsForQuality(hotel model.HotelOffer) []string {
	text := strings.ToLower(strings.Join([]string{
		hotel.Name,
		hotel.Address,
		hotel.Star,
		hotel.NearbyPOI,
	}, " "))
	tags := make([]string, 0)

	if containsAnyText(text, []string{"青旅", "青年旅舍", "多人间", "床位房", "hostel"}) {
		tags = append(tags, "hostel")
	}
	if containsAnyText(text, []string{"民宿"}) {
		tags = append(tags, "homestay")
	}
	if containsAnyText(text, []string{"客栈", "旅社"}) {
		tags = append(tags, "guesthouse")
	}
	if containsAnyText(text, []string{"公寓"}) {
		tags = append(tags, "apartment")
	}

	return uniqueStringList(tags)
}

func hotelDistanceToRouteCore(hotel model.HotelOffer, dailyRoutes []model.DailyRoute) float64 {
	if hotel.Lat == 0 && hotel.Lng == 0 {
		return 0
	}

	lat, lng, ok := routeCoreCenterInTools(dailyRoutes)
	if !ok {
		return 0
	}

	return math.Round(distanceKM(hotel.Lat, hotel.Lng, lat, lng)*10) / 10
}

func routeCoreCenterInTools(dailyRoutes []model.DailyRoute) (float64, float64, bool) {
	mainPoints := make([]model.Attraction, 0)
	fallbackPoints := make([]model.Attraction, 0)

	for _, route := range dailyRoutes {
		for _, attraction := range route.Attractions {
			if attraction.Lat == 0 && attraction.Lng == 0 {
				continue
			}

			fallbackPoints = append(fallbackPoints, attraction)
			if BuildTripPOI(attraction, 0).Role == tripPOIRoleMainAttraction {
				mainPoints = append(mainPoints, attraction)
			}
		}
	}

	if len(mainPoints) > 0 {
		return attractionCenterInTools(mainPoints)
	}

	return attractionCenterInTools(fallbackPoints)
}

func attractionCenterInTools(attractions []model.Attraction) (float64, float64, bool) {
	if len(attractions) == 0 {
		return 0, 0, false
	}

	lat := 0.0
	lng := 0.0

	for _, attraction := range attractions {
		lat += attraction.Lat
		lng += attraction.Lng
	}

	count := float64(len(attractions))
	return lat / count, lng / count, true
}
