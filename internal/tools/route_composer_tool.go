package tools

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"ariadne/internal/model"
)

const (
	tripPOIRoleMainAttraction = "main_attraction"
	tripPOIRoleFoodSpot       = "food_spot"
	tripPOIRoleShoppingSpot   = "shopping_spot"
	tripPOIRoleTransitNearby  = "transit_nearby"
	tripPOIRoleInvalid        = "invalid"
)

func BuildTripPOIs(attractions []model.Attraction) []model.TripPOI {
	pois := make([]model.TripPOI, 0, len(attractions))

	for i, attraction := range attractions {
		pois = append(pois, BuildTripPOI(attraction, i))
	}

	sort.SliceStable(pois, func(i, j int) bool {
		if pois[i].Score != pois[j].Score {
			return pois[i].Score > pois[j].Score
		}

		return pois[i].Name < pois[j].Name
	})

	return pois
}

func BuildTripPOI(attraction model.Attraction, rank int) model.TripPOI {
	profile := BuildPOIProfile(attraction)
	tags := enrichTripPOITags(profile.Tags, attraction)
	role := classifyTripPOIRole(profile, tags)

	poi := model.TripPOI{
		Name:        attraction.Name,
		Category:    attraction.Category,
		Address:     attraction.Address,
		Description: attraction.Description,
		Lat:         attraction.Lat,
		Lng:         attraction.Lng,
		Role:        role,
		Tags:        tags,
		Score:       scoreTripPOI(tags, role, rank),
		Attraction:  attraction,
	}

	if profile.Invalid {
		poi.Role = tripPOIRoleInvalid
		poi.Score = -1000
	}

	return poi
}

func ComposeDailyRoutes(
	request model.TripRequest,
	attractions []model.Attraction,
	preferenceProfile model.EffectivePreferenceProfile,
) []model.DailyRoute {
	if request.Days <= 0 {
		request.Days = 1
	}

	pois := BuildTripPOIs(attractions)
	return ComposeDailyRoutesFromTripPOIs(request, pois, preferenceProfile)
}

func ComposeDailyRoutesFromTripPOIs(
	request model.TripRequest,
	pois []model.TripPOI,
	preferenceProfile model.EffectivePreferenceProfile,
) []model.DailyRoute {
	if request.Days <= 0 {
		request.Days = 1
	}

	clusters := ClusterTripPOIs(pois)
	routes := make([]model.DailyRoute, 0, request.Days)
	usedPOIs := map[string]bool{}
	usedClusters := map[int]bool{}

	for day := 1; day <= request.Days; day++ {
		template := buildDayPlanTemplate(day, request.Days, request, preferenceProfile)
		clusterIndex := pickClusterForDay(clusters, usedClusters, usedPOIs)
		selected := selectPOIsForDay(template, clusterIndex, clusters, pois, usedPOIs, preferenceProfile)

		if clusterIndex >= 0 {
			usedClusters[clusterIndex] = true
		}

		for _, poi := range selected {
			usedPOIs[tripPOIKey(poi)] = true
		}

		attractionsForDay := tripPOIsToAttractions(selected)
		route := model.DailyRoute{
			Day:           day,
			Title:         fmt.Sprintf("Day %d %s route", day, template.DayType),
			Attractions:   attractionsForDay,
			Summary:       buildRouteComposerSummary(template, selected),
			EstimatedCost: calculateRouteCost(attractionsForDay),
			DataSource:    "route_composer",
		}

		routes = append(routes, route)
	}

	return routes
}

func ClusterTripPOIs(pois []model.TripPOI) []model.POICluster {
	const clusterRadiusKM = 5.0

	clusters := make([]model.POICluster, 0)
	noCoordinatePOIs := make([]model.TripPOI, 0)

	for _, poi := range pois {
		if poi.Role == tripPOIRoleInvalid {
			continue
		}

		if !hasPOICoordinates(poi) {
			noCoordinatePOIs = append(noCoordinatePOIs, poi)
			continue
		}

		bestIndex := -1
		bestDistance := 0.0
		for i, cluster := range clusters {
			if cluster.CenterLat == 0 && cluster.CenterLng == 0 {
				continue
			}

			distance := distanceKM(poi.Lat, poi.Lng, cluster.CenterLat, cluster.CenterLng)
			if distance <= clusterRadiusKM && (bestIndex == -1 || distance < bestDistance) {
				bestIndex = i
				bestDistance = distance
			}
		}

		if bestIndex == -1 {
			clusters = append(clusters, newPOICluster(poi))
			continue
		}

		addPOIToCluster(&clusters[bestIndex], poi)
	}

	if len(noCoordinatePOIs) > 0 {
		cluster := model.POICluster{
			POIs: noCoordinatePOIs,
		}
		for _, poi := range noCoordinatePOIs {
			cluster.Score += poi.Score
			cluster.Tags = append(cluster.Tags, poi.Tags...)
		}
		cluster.Tags = uniqueStringList(cluster.Tags)
		clusters = append(clusters, cluster)
	}

	sort.SliceStable(clusters, func(i, j int) bool {
		if clusters[i].Score != clusters[j].Score {
			return clusters[i].Score > clusters[j].Score
		}

		return len(clusters[i].POIs) > len(clusters[j].POIs)
	})

	return clusters
}

func classifyTripPOIRole(profile model.POIProfile, tags []string) string {
	if profile.Invalid {
		return tripPOIRoleInvalid
	}

	if hasAnyTag(tags, []string{"food", "local_food", "snack_street"}) &&
		!hasAnyTag(tags, []string{"old_street", "historic_site", "landmark", "night_view", "waterfront"}) {
		return tripPOIRoleFoodSpot
	}

	if hasAnyTag(tags, []string{"shopping", "commercial_area"}) &&
		!hasAnyTag(tags, []string{"old_street", "historic_site", "landmark", "night_view", "waterfront"}) {
		return tripPOIRoleShoppingSpot
	}

	if hasAnyTag(tags, []string{
		"museum",
		"exhibition",
		"art_gallery",
		"memorial",
		"science_museum",
		"aquarium",
		"zoo",
		"historic_site",
		"old_street",
		"city_walk",
		"park",
		"nature",
		"landmark",
		"night_view",
		"waterfront",
		"sea",
		"beach",
		"coast",
	}) {
		return tripPOIRoleMainAttraction
	}

	if hasAnyTag(tags, []string{"food", "local_food", "snack_street"}) {
		return tripPOIRoleFoodSpot
	}

	if hasAnyTag(tags, []string{"shopping", "commercial_area"}) {
		return tripPOIRoleShoppingSpot
	}

	if hasAnyTag(tags, []string{"station", "airport", "transit"}) {
		return tripPOIRoleTransitNearby
	}

	return tripPOIRoleMainAttraction
}

func enrichTripPOITags(tags []string, attraction model.Attraction) []string {
	result := append([]string{}, tags...)
	text := strings.ToLower(strings.Join([]string{
		attraction.Name,
		attraction.Category,
		attraction.Address,
		attraction.Description,
	}, " "))

	if containsAnyText(text, []string{"老街", "古街", "历史街区", "步行街", "街区", "胡同", "巷子"}) {
		result = append(result, "old_street", "city_walk", "historic_site")
	}

	if containsAnyText(text, []string{"小吃街", "美食街", "夜市", "老字号"}) {
		result = append(result, "food", "local_food", "snack_street")
	}

	if containsAnyText(text, []string{"夜景", "观景", "观景台", "滨江", "江畔", "河畔", "湖滨", "外滩", "码头"}) {
		result = append(result, "night_view", "landmark", "waterfront")
	}

	if containsAnyText(text, []string{"看海", "海边", "海滩", "沙滩", "海滨", "海岸", "滨海", "观海", "海湾", "海堤", "栈道"}) {
		result = append(result, "sea", "beach", "waterfront", "coast", "low_exertion")
	}

	if containsAnyText(text, []string{"商圈", "商业区", "购物中心", "商场", "市场"}) {
		result = append(result, "shopping", "commercial_area")
	}

	if containsAnyText(text, []string{"科技馆", "科学中心", "科普馆"}) {
		result = append(result, "science_museum", "family", "indoor")
	}

	if containsAnyText(text, []string{"海洋馆", "水族馆"}) {
		result = append(result, "aquarium", "family", "indoor")
	}

	return uniqueStringList(result)
}

func scoreTripPOI(tags []string, role string, rank int) int {
	score := 80 - rank
	if score < 0 {
		score = 0
	}

	switch role {
	case tripPOIRoleMainAttraction:
		score += 100
	case tripPOIRoleFoodSpot:
		score += 55
	case tripPOIRoleShoppingSpot:
		score += 40
	case tripPOIRoleTransitNearby:
		score += 10
	default:
		score -= 100
	}

	if hasAnyTag(tags, []string{"museum", "science_museum", "aquarium", "historic_site", "old_street", "night_view", "landmark", "sea", "beach", "waterfront", "coast"}) {
		score += 20
	}

	if hasAnyTag(tags, []string{"family", "indoor", "low_exertion", "city_walk"}) {
		score += 8
	}

	if hasAnyTag(tags, []string{"high_exertion", "mountain", "hiking", "remote"}) {
		score -= 25
	}

	return score
}

func buildDayPlanTemplate(
	day int,
	totalDays int,
	request model.TripRequest,
	preferenceProfile model.EffectivePreferenceProfile,
) model.DayPlanTemplate {
	dayType := "full_day"
	if totalDays == 1 {
		dayType = "full_day"
	} else if day == 1 {
		dayType = "arrival"
	} else if day == totalDays {
		dayType = "departure"
	}

	template := model.DayPlanTemplate{
		DayType:            dayType,
		MainPOICount:       2,
		FoodPOICount:       1,
		ShoppingPOICount:   0,
		AllowNightView:     dayType != "departure",
		MaxTransferMinutes: 60,
	}

	if dayType == "arrival" || dayType == "departure" {
		template.MainPOICount = 1
		template.MaxTransferMinutes = 45
	}

	if isRelaxedOrChildFriendlyTrip(request, preferenceProfile) {
		template.MaxTransferMinutes = 35
		if dayType == "full_day" {
			template.MainPOICount = 2
		} else {
			template.MainPOICount = 1
		}
	}

	if isIntensiveTrip(request, preferenceProfile) {
		template.MaxTransferMinutes = 90
		if dayType == "full_day" {
			template.MainPOICount = 4
		} else {
			template.MainPOICount = 2
		}
	}

	if prefersShopping(preferenceProfile) {
		template.ShoppingPOICount = 1
	}

	return template
}

func pickClusterForDay(
	clusters []model.POICluster,
	usedClusters map[int]bool,
	usedPOIs map[string]bool,
) int {
	bestIndex := -1
	bestScore := 0

	for i, cluster := range clusters {
		score := cluster.Score + unusedRoleCount(cluster.POIs, usedPOIs, tripPOIRoleMainAttraction)*100
		score += unusedRoleCount(cluster.POIs, usedPOIs, tripPOIRoleFoodSpot) * 10

		if usedClusters[i] {
			score -= 60
		}

		if !clusterHasUnusedPOIs(cluster, usedPOIs) {
			continue
		}

		if bestIndex == -1 || score > bestScore {
			bestIndex = i
			bestScore = score
		}
	}

	return bestIndex
}

func selectPOIsForDay(
	template model.DayPlanTemplate,
	clusterIndex int,
	clusters []model.POICluster,
	allPOIs []model.TripPOI,
	usedPOIs map[string]bool,
	preferenceProfile model.EffectivePreferenceProfile,
) []model.TripPOI {
	clusterPOIs := []model.TripPOI{}
	if clusterIndex >= 0 && clusterIndex < len(clusters) {
		clusterPOIs = clusters[clusterIndex].POIs
	}

	selected := make([]model.TripPOI, 0)
	selected = appendRequiredIntentPOIs(selected, clusterPOIs, usedPOIs, preferenceProfile)
	if !selectedSatisfiesRequiredIntent(selected, preferenceProfile) {
		selected = appendRequiredIntentPOIs(selected, allPOIs, usedPOIs, preferenceProfile)
	}

	selected = appendSelectedPOIs(selected, clusterPOIs, usedPOIs, tripPOIRoleMainAttraction, template.MainPOICount)

	if countRole(selected, tripPOIRoleMainAttraction) < template.MainPOICount {
		remaining := template.MainPOICount - countRole(selected, tripPOIRoleMainAttraction)
		selected = appendSelectedPOIs(selected, allPOIs, usedPOIs, tripPOIRoleMainAttraction, remaining)
	}

	if template.AllowNightView &&
		prefersNightView(preferenceProfile) &&
		!selectedHasTag(selected, "night_view") &&
		len(selected) < template.MainPOICount+1 {
		selected = appendSelectedTaggedPOI(selected, clusterPOIs, usedPOIs, "night_view")
		if !selectedHasTag(selected, "night_view") {
			selected = appendSelectedTaggedPOI(selected, allPOIs, usedPOIs, "night_view")
		}
	}

	if template.FoodPOICount > 0 && (prefersFood(preferenceProfile) || countRole(selected, tripPOIRoleMainAttraction) > 0) {
		selected = appendSelectedPOIs(selected, clusterPOIs, usedPOIs, tripPOIRoleFoodSpot, template.FoodPOICount)
		if countRole(selected, tripPOIRoleFoodSpot) == 0 {
			selected = appendSelectedPOIs(selected, allPOIs, usedPOIs, tripPOIRoleFoodSpot, template.FoodPOICount)
		}
	}

	if template.ShoppingPOICount > 0 {
		selected = appendSelectedPOIs(selected, clusterPOIs, usedPOIs, tripPOIRoleShoppingSpot, template.ShoppingPOICount)
		if countRole(selected, tripPOIRoleShoppingSpot) == 0 {
			selected = appendSelectedPOIs(selected, allPOIs, usedPOIs, tripPOIRoleShoppingSpot, template.ShoppingPOICount)
		}
	}

	if len(selected) == 0 {
		selected = appendSelectedPOIs(selected, allPOIs, usedPOIs, tripPOIRoleMainAttraction, 1)
	}

	if len(selected) == 0 && !hasUnusedRole(allPOIs, usedPOIs, tripPOIRoleMainAttraction) {
		selected = appendSelectedPOIs(selected, allPOIs, usedPOIs, tripPOIRoleFoodSpot, 1)
	}

	return selected
}

func appendSelectedPOIs(
	selected []model.TripPOI,
	candidates []model.TripPOI,
	usedPOIs map[string]bool,
	role string,
	limit int,
) []model.TripPOI {
	if limit <= 0 {
		return selected
	}

	sorted := sortedPOICandidates(candidates)
	added := 0
	selectedKeys := selectedPOIKeys(selected)

	for _, poi := range sorted {
		if poi.Role != role || poi.Role == tripPOIRoleInvalid {
			continue
		}

		key := tripPOIKey(poi)
		if usedPOIs[key] || selectedKeys[key] {
			continue
		}

		selected = append(selected, poi)
		selectedKeys[key] = true
		added++

		if added >= limit {
			break
		}
	}

	return selected
}

func appendSelectedTaggedPOI(
	selected []model.TripPOI,
	candidates []model.TripPOI,
	usedPOIs map[string]bool,
	tag string,
) []model.TripPOI {
	sorted := sortedPOICandidates(candidates)
	selectedKeys := selectedPOIKeys(selected)

	for _, poi := range sorted {
		key := tripPOIKey(poi)
		if poi.Role == tripPOIRoleInvalid || usedPOIs[key] || selectedKeys[key] {
			continue
		}

		if hasTag(poi.Tags, tag) {
			return append(selected, poi)
		}
	}

	return selected
}

func appendRequiredIntentPOIs(
	selected []model.TripPOI,
	candidates []model.TripPOI,
	usedPOIs map[string]bool,
	preferenceProfile model.EffectivePreferenceProfile,
) []model.TripPOI {
	requiredTags := requiredRouteIntentTags(preferenceProfile)
	if len(requiredTags) == 0 || selectedHasAnyTag(selected, requiredTags) {
		return selected
	}

	sorted := sortedPOICandidates(candidates)
	selectedKeys := selectedPOIKeys(selected)

	for _, poi := range sorted {
		key := tripPOIKey(poi)
		if poi.Role == tripPOIRoleInvalid || usedPOIs[key] || selectedKeys[key] {
			continue
		}

		if hasAnyTag(poi.Tags, requiredTags) {
			return append(selected, poi)
		}
	}

	return selected
}

func selectedSatisfiesRequiredIntent(pois []model.TripPOI, preferenceProfile model.EffectivePreferenceProfile) bool {
	requiredTags := requiredRouteIntentTags(preferenceProfile)
	return len(requiredTags) == 0 || selectedHasAnyTag(pois, requiredTags)
}

func requiredRouteIntentTags(preferenceProfile model.EffectivePreferenceProfile) []string {
	for _, tag := range []string{"sea", "beach", "waterfront", "coast"} {
		if hasString(preferenceProfile.Attraction.HardPreferTags, tag) || hasString(preferenceProfile.Attraction.SoftPreferTags, tag) {
			return []string{"sea", "beach", "waterfront", "coast"}
		}
	}

	return nil
}

func sortedPOICandidates(candidates []model.TripPOI) []model.TripPOI {
	result := append([]model.TripPOI{}, candidates...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Score != result[j].Score {
			return result[i].Score > result[j].Score
		}

		return result[i].Name < result[j].Name
	})

	return result
}

func newPOICluster(poi model.TripPOI) model.POICluster {
	return model.POICluster{
		CenterLat: poi.Lat,
		CenterLng: poi.Lng,
		POIs:      []model.TripPOI{poi},
		Tags:      uniqueStringList(poi.Tags),
		Score:     poi.Score + roleClusterBoost(poi.Role),
	}
}

func addPOIToCluster(cluster *model.POICluster, poi model.TripPOI) {
	count := len(cluster.POIs)
	cluster.CenterLat = (cluster.CenterLat*float64(count) + poi.Lat) / float64(count+1)
	cluster.CenterLng = (cluster.CenterLng*float64(count) + poi.Lng) / float64(count+1)
	cluster.POIs = append(cluster.POIs, poi)
	cluster.Tags = uniqueStringList(append(cluster.Tags, poi.Tags...))
	cluster.Score += poi.Score + roleClusterBoost(poi.Role)
}

func roleClusterBoost(role string) int {
	switch role {
	case tripPOIRoleMainAttraction:
		return 80
	case tripPOIRoleFoodSpot:
		return 20
	case tripPOIRoleShoppingSpot:
		return 10
	default:
		return 0
	}
}

func tripPOIsToAttractions(pois []model.TripPOI) []model.Attraction {
	attractions := make([]model.Attraction, 0, len(pois))

	for _, poi := range pois {
		attractions = append(attractions, poi.Attraction)
	}

	return attractions
}

func buildRouteComposerSummary(template model.DayPlanTemplate, pois []model.TripPOI) string {
	mainCount := countRole(pois, tripPOIRoleMainAttraction)
	foodCount := countRole(pois, tripPOIRoleFoodSpot)

	if mainCount == 0 && foodCount > 0 {
		return "Route quality warning: this day only has food spots because no usable main attraction was available."
	}

	if foodCount > 0 {
		return fmt.Sprintf("Composed as a %s day with %d main attraction(s) and nearby food support.", template.DayType, mainCount)
	}

	return fmt.Sprintf("Composed as a %s day with %d main attraction(s).", template.DayType, mainCount)
}

func tripPOIKey(poi model.TripPOI) string {
	key := strings.ToLower(strings.TrimSpace(poi.Name) + "|" + strings.TrimSpace(poi.Address))
	if key == "|" {
		key = strings.ToLower(strings.TrimSpace(poi.Attraction.Link))
	}

	return key
}

func selectedPOIKeys(pois []model.TripPOI) map[string]bool {
	result := map[string]bool{}

	for _, poi := range pois {
		result[tripPOIKey(poi)] = true
	}

	return result
}

func clusterHasUnusedPOIs(cluster model.POICluster, usedPOIs map[string]bool) bool {
	for _, poi := range cluster.POIs {
		if !usedPOIs[tripPOIKey(poi)] {
			return true
		}
	}

	return false
}

func unusedRoleCount(pois []model.TripPOI, usedPOIs map[string]bool, role string) int {
	count := 0

	for _, poi := range pois {
		if poi.Role == role && !usedPOIs[tripPOIKey(poi)] {
			count++
		}
	}

	return count
}

func countRole(pois []model.TripPOI, role string) int {
	count := 0

	for _, poi := range pois {
		if poi.Role == role {
			count++
		}
	}

	return count
}

func hasUnusedRole(pois []model.TripPOI, usedPOIs map[string]bool, role string) bool {
	for _, poi := range pois {
		if poi.Role == role && !usedPOIs[tripPOIKey(poi)] {
			return true
		}
	}

	return false
}

func selectedHasTag(pois []model.TripPOI, tag string) bool {
	for _, poi := range pois {
		if hasTag(poi.Tags, tag) {
			return true
		}
	}

	return false
}

func selectedHasAnyTag(pois []model.TripPOI, tags []string) bool {
	for _, tag := range tags {
		if selectedHasTag(pois, tag) {
			return true
		}
	}

	return false
}

func hasPOICoordinates(poi model.TripPOI) bool {
	return poi.Lat != 0 || poi.Lng != 0
}

func distanceKM(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKM = 6371.0

	dLat := degreesToRadians(lat2 - lat1)
	dLng := degreesToRadians(lng2 - lng1)
	rLat1 := degreesToRadians(lat1)
	rLat2 := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKM * c
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func isRelaxedOrChildFriendlyTrip(
	request model.TripRequest,
	preferenceProfile model.EffectivePreferenceProfile,
) bool {
	text := strings.ToLower(request.RawInput + " " + request.Preference)

	return containsAnyText(text, []string{"轻松", "慢节奏", "亲子", "带孩子", "带小孩", "儿童"}) ||
		hasString(preferenceProfile.Route.HardPreferTags, "relaxed") ||
		hasString(preferenceProfile.Route.SoftPreferTags, "relaxed") ||
		hasString(preferenceProfile.Route.HardPreferTags, "child_friendly") ||
		hasString(preferenceProfile.Route.SoftPreferTags, "child_friendly") ||
		hasString(preferenceProfile.Attraction.HardPreferTags, "family") ||
		hasString(preferenceProfile.Attraction.SoftPreferTags, "family")
}

func isIntensiveTrip(
	request model.TripRequest,
	preferenceProfile model.EffectivePreferenceProfile,
) bool {
	text := strings.ToLower(request.RawInput + " " + request.Preference)

	return containsAnyText(text, []string{"特种兵", "多玩", "尽量多", "行程紧"}) ||
		hasString(preferenceProfile.Route.HardPreferTags, "intensive") ||
		hasString(preferenceProfile.Route.SoftPreferTags, "intensive")
}

func prefersFood(preferenceProfile model.EffectivePreferenceProfile) bool {
	return hasString(preferenceProfile.Food.HardPreferTags, "food") ||
		hasString(preferenceProfile.Food.SoftPreferTags, "food") ||
		hasString(preferenceProfile.Food.HardPreferTags, "local_food") ||
		hasString(preferenceProfile.Food.SoftPreferTags, "local_food") ||
		hasString(preferenceProfile.Attraction.HardPreferTags, "local_food") ||
		hasString(preferenceProfile.Attraction.SoftPreferTags, "local_food")
}

func prefersShopping(preferenceProfile model.EffectivePreferenceProfile) bool {
	return hasString(preferenceProfile.Attraction.HardPreferTags, "shopping") ||
		hasString(preferenceProfile.Attraction.SoftPreferTags, "shopping") ||
		hasString(preferenceProfile.Attraction.HardPreferTags, "commercial_area") ||
		hasString(preferenceProfile.Attraction.SoftPreferTags, "commercial_area")
}

func prefersNightView(preferenceProfile model.EffectivePreferenceProfile) bool {
	return hasString(preferenceProfile.Attraction.HardPreferTags, "night_view") ||
		hasString(preferenceProfile.Attraction.SoftPreferTags, "night_view")
}
