package tools

import (
	"math"
	"strings"

	"ariadne/internal/model"
)

func BuildPlanQualityReport(
	dailyRoutes []model.DailyRoute,
	attractions []model.Attraction,
	hotel model.HotelOffer,
	violations []model.RecommendationViolation,
	preferenceProfile model.EffectivePreferenceProfile,
) model.PlanQualityReport {
	report := model.PlanQualityReport{
		HardConstraintPassed: true,
		Warnings:             make([]string, 0),
		Score:                100,
	}

	for _, violation := range violations {
		if strings.Contains(strings.ToLower(violation.Type), "hard_constraint") {
			report.HardConstraintPassed = false
			report.Warnings = append(report.Warnings, violation.Message)
			report.Score -= 40
		}
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

	requiredTags := requiredRouteIntentTags(preferenceProfile)
	if len(requiredTags) > 0 && !dailyRoutesContainAnyTag(dailyRoutes, requiredTags) {
		report.Warnings = append(report.Warnings, "Daily routes do not satisfy a core attraction intent from the current request.")
		report.Score -= 20
	}

	if report.FoodSpotCount > report.MainAttractionCount && report.MainAttractionCount > 0 {
		report.Warnings = append(report.Warnings, "Food spots outnumber main attractions; route composition may be too food-heavy.")
		report.Score -= 8
	}

	if report.InvalidPOICount > 0 {
		report.Warnings = append(report.Warnings, "Invalid or non-travel POIs were detected in the candidate set.")
		report.Score -= report.InvalidPOICount * 5
	}

	if report.AvgTransferMinutes > 60 {
		report.Warnings = append(report.Warnings, "Average transfer time is high.")
		report.Score -= 10
	}

	report.HotelDistanceToCoreKM = hotelDistanceToRouteCore(hotel, dailyRoutes)
	if report.HotelDistanceToCoreKM > 15 {
		report.Warnings = append(report.Warnings, "Recommended hotel is far from the route core.")
		report.Score -= 15
	} else if report.HotelDistanceToCoreKM > 8 {
		report.Warnings = append(report.Warnings, "Recommended hotel is not close to the route core.")
		report.Score -= 8
	}

	if report.Score < 0 {
		report.Score = 0
	}

	return report
}

func dailyRoutesContainAnyTag(dailyRoutes []model.DailyRoute, tags []string) bool {
	for _, route := range dailyRoutes {
		for _, attraction := range route.Attractions {
			poi := BuildTripPOI(attraction, 0)
			if hasAnyTag(poi.Tags, tags) {
				return true
			}
		}
	}

	return false
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
