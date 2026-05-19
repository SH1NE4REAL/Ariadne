package tools

import "ariadne/internal/model"

func BuildPOIDebugReport(
	rawAttractions []model.Attraction,
	finalAttractions []model.Attraction,
	fallbackKeywords []string,
	rawPOICountByKeyword map[string]int,
	searchStatusByKeyword map[string]string,
	searchErrorByKeyword map[string]string,
	preferenceProfile model.EffectivePreferenceProfile,
) model.POIDebugReport {
	report := model.POIDebugReport{
		RawPOICount:          len(rawAttractions),
		RawPOICountByKeyword: cloneIntMap(rawPOICountByKeyword),
		RejectedReasons:      map[string]int{},
		FallbackKeywords:     uniqueStrings(fallbackKeywords),
		SearchStatusByKeyword: cloneStringMap(
			searchStatusByKeyword,
		),
		SearchErrorByKeyword: cloneStringMap(searchErrorByKeyword),
	}

	seenFinal := map[string]bool{}
	for _, attraction := range finalAttractions {
		seenFinal[debugAttractionKey(attraction)] = true
	}

	for _, attraction := range rawAttractions {
		poi := BuildTripPOI(attraction, 0)

		if poi.Role == tripPOIRoleInvalid {
			report.RejectedReasons["invalid_poi"]++
			continue
		}

		if poi.Role == tripPOIRoleSupportSpot || poi.Role == tripPOIRoleTransitNearby {
			report.RejectedReasons[poi.Role]++
			continue
		}

		report.AfterRoleClassifyCount++

		if reason := poiHardAvoidRejectReason(poi, preferenceProfile); reason != "" {
			report.RejectedReasons[reason]++
			continue
		}

		report.AfterHardAvoidFilterCount++
		report.AfterDistanceFilterCount++

		if seenFinal[debugAttractionKey(attraction)] {
			report.FinalRoutablePOICount++
		}
	}

	return report
}

func cloneIntMap(items map[string]int) map[string]int {
	result := map[string]int{}
	for key, value := range items {
		result[key] = value
	}

	return result
}

func cloneStringMap(items map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range items {
		result[key] = value
	}

	return result
}

func poiHardAvoidRejectReason(poi model.TripPOI, preferenceProfile model.EffectivePreferenceProfile) string {
	if hardAvoidsShopping(preferenceProfile) &&
		(poi.Role == tripPOIRoleShoppingSpot || hasAnyTag(poi.Tags, []string{"shopping", "commercial_area"})) {
		return "hard_avoid_shopping"
	}

	if (hasString(preferenceProfile.Attraction.HardAvoidTags, "mountain") ||
		hasString(preferenceProfile.Attraction.HardAvoidTags, "hiking") ||
		hasString(preferenceProfile.Attraction.HardAvoidTags, "high_exertion")) &&
		isMountainLikePOI(poi) {
		return "hard_avoid_mountain"
	}

	for _, tag := range preferenceProfile.Attraction.HardAvoidTags {
		if hasTag(poi.Tags, tag) {
			return "hard_avoid_" + tag
		}
	}

	return ""
}

func debugAttractionKey(attraction model.Attraction) string {
	return attraction.Name + "|" + attraction.Address + "|" + attraction.Category
}
