package tools

import (
	"sort"
	"strings"

	"ariadne/internal/model"
)

type PreferenceAdapterTool struct {
	Name        string
	Description string
}

func NewPreferenceAdapterTool() PreferenceAdapterTool {
	return PreferenceAdapterTool{
		Name:        "preference_adapter_tool",
		Description: "将结构化偏好约束应用到交通、景点和路线推荐中",
	}
}

func (t PreferenceAdapterTool) FilterTrainOffers(
	request model.TripRequest,
	offers []model.TrainOffer,
	constraints []model.PreferenceConstraint,
) []model.TrainOffer {
	if len(offers) == 0 {
		return offers
	}

	profile := ResolvePreferenceConstraints(constraints)
	requireHighSpeed := hasString(profile.Transport.HardPreferTags, "high_speed_train") ||
		hasString(profile.Transport.HardPreferTags, "bullet_train")

	candidates := make([]model.TrainOffer, 0)

	for _, offer := range offers {
		text := buildTrainConstraintText(offer)

		if currentRequestRequiresHighSpeed(request) {
			if strings.Contains(text, "普快") || strings.Contains(text, "硬座") {
				continue
			}
		}

		if requireHighSpeed && !isHighSpeedTrainOfferInAdapter(offer) {
			continue
		}

		if violatesHardConstraint(text, constraints, "transport") {
			continue
		}

		candidates = append(candidates, offer)
	}

	if len(candidates) == 0 {
		if requireHighSpeed || currentRequestRequiresHighSpeed(request) {
			return []model.TrainOffer{}
		}
		return offers
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		scoreI := scoreTextByConstraints(buildTrainConstraintText(candidates[i]), constraints, "transport")
		scoreJ := scoreTextByConstraints(buildTrainConstraintText(candidates[j]), constraints, "transport")

		if scoreI != scoreJ {
			return scoreI > scoreJ
		}

		return candidates[i].Price < candidates[j].Price
	})

	return candidates
}

func (t PreferenceAdapterTool) FilterFlightOffers(
	request model.TripRequest,
	offers []model.FlightOffer,
	constraints []model.PreferenceConstraint,
) []model.FlightOffer {
	if len(offers) == 0 {
		return offers
	}

	profile := ResolvePreferenceConstraints(constraints)
	if hasString(profile.Transport.HardAvoidTags, "flight") {
		return []model.FlightOffer{}
	}

	if currentRequestRequiresFlight(request) {
		return offers
	}

	candidates := make([]model.FlightOffer, 0)

	for _, offer := range offers {
		text := buildFlightConstraintText(offer)

		if violatesHardConstraint(text, constraints, "transport") {
			continue
		}

		candidates = append(candidates, offer)
	}

	if len(candidates) == 0 {
		return offers
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		scoreI := scoreTextByConstraints(buildFlightConstraintText(candidates[i]), constraints, "transport")
		scoreJ := scoreTextByConstraints(buildFlightConstraintText(candidates[j]), constraints, "transport")

		if scoreI != scoreJ {
			return scoreI > scoreJ
		}

		return candidates[i].Price < candidates[j].Price
	})

	return candidates
}

func isHighSpeedTrainOfferInAdapter(offer model.TrainOffer) bool {
	for _, segment := range offer.Segments {
		if strings.Contains(segment.TrainType, "高铁") ||
			strings.Contains(segment.TrainType, "动车") ||
			strings.Contains(segment.TrainType, "城际") {
			return true
		}

		if len(segment.TrainNo) > 0 {
			first := segment.TrainNo[0]
			if first == 'G' || first == 'D' || first == 'C' {
				return true
			}
		}
	}

	return false
}

func buildTrainConstraintText(offer model.TrainOffer) string {
	parts := []string{
		offer.JourneyType,
		offer.Direction,
	}

	for _, segment := range offer.Segments {
		parts = append(parts,
			segment.TrainNo,
			segment.TrainType,
			segment.SeatClassName,
			segment.DepStationName,
			segment.ArrStationName,
		)
	}

	return strings.ToLower(strings.Join(parts, " "))
}

func buildFlightConstraintText(offer model.FlightOffer) string {
	parts := []string{
		"飞机",
		"航班",
		offer.Provider,
	}

	for _, journey := range offer.Journeys {
		parts = append(parts, journey.Direction, journey.JourneyType)

		for _, segment := range journey.Segments {
			parts = append(parts,
				segment.FlightNo,
				segment.Airline,
				segment.SeatClassName,
				segment.DepAirportName,
				segment.ArrAirportName,
			)
		}
	}

	return strings.ToLower(strings.Join(parts, " "))
}

func currentRequestRequiresFlight(request model.TripRequest) bool {
	text := strings.ToLower(request.TransportPreference + " " + request.Preference + " " + request.RawInput)
	return strings.Contains(text, "飞机") || strings.Contains(text, "航班")
}

func currentRequestRequiresHighSpeed(request model.TripRequest) bool {
	text := strings.ToLower(request.TransportPreference + " " + request.Preference + " " + request.RawInput)
	return strings.Contains(text, "高铁") || strings.Contains(text, "动车")
}

func (t PreferenceAdapterTool) BuildAttractionSearchKeywords(
	request model.TripRequest,
	constraints []model.PreferenceConstraint,
) []string {
	text := strings.ToLower(request.RawInput + " " + request.Preference)
	keywords := make([]string, 0)
	currentIntent := BuildTripIntentProfile(constraints, "current_request")
	searchConstraints := constraintsForPOISearch(constraints)
	if !TripIntentProfileHasPOISearchIntent(currentIntent) {
		searchConstraints = constraints
	}

	profile := ResolvePreferenceConstraints(searchConstraints)
	allProfile := ResolvePreferenceConstraints(constraints)
	attractionPreference := profile.Attraction
	foodPreference := profile.Food
	avoidMuseum := hasEffectiveHardAvoidTag(allProfile.Attraction, "museum") ||
		hasEffectiveHardAvoidTag(allProfile.Attraction, "exhibition") ||
		hasEffectiveHardAvoidTag(allProfile.Attraction, "art_gallery")
	avoidShopping := hasEffectiveHardAvoidTag(allProfile.Attraction, "shopping") ||
		hasEffectiveHardAvoidTag(allProfile.Attraction, "commercial_area")
	avoidMountain := hasEffectiveHardAvoidTag(allProfile.Attraction, "mountain") ||
		hasEffectiveHardAvoidTag(allProfile.Attraction, "hiking") ||
		hasEffectiveHardAvoidTag(allProfile.Attraction, "high_exertion")

	if isSameDayBusinessTrip(request, allProfile) {
		return []string{"城市地标", "市区漫步"}
	}

	if !avoidMuseum && containsAnyText(text, []string{"博物馆", "展览", "美术馆", "人文", "历史"}) {
		keywords = append(keywords, "博物馆", "展览馆", "美术馆")
	}

	if containsAnyText(text, []string{"小吃", "美食", "地道吃的", "本地吃的"}) {
		keywords = append(keywords, "小吃街", "美食街")
	}

	if containsAnyText(text, []string{"夜景", "晚上逛", "夜游"}) {
		keywords = append(keywords, "夜景", "观景")
	}

	if containsAnyText(text, []string{"老街", "市区老街", "历史街区", "逛街"}) {
		keywords = append(keywords, "老街", "历史街区")
	}

	if containsAnyText(text, []string{"看海", "海边", "海滩", "沙滩", "海滨", "海岸", "滨海", "观海"}) {
		keywords = append([]string{"海边", "海滩"}, keywords...)
	}

	if containsAnyText(text, []string{"科技馆", "科学中心", "海洋馆", "水族馆", "科普"}) {
		keywords = append(keywords, "科技馆", "海洋馆", "水族馆", "儿童展览", "科普", "科学中心", "自然博物馆")
	}

	if containsAnyText(text, []string{"亲子", "带孩子", "带小孩"}) {
		keywords = append(keywords, "科技馆", "海洋馆", "动物园")
	}

	if containsAnyText(text, []string{"购物", "商场", "买东西", "逛街"}) {
		keywords = append(keywords, "商场", "购物中心", "商业街")
	}

	if containsAnyText(text, []string{"自然风景", "看风景", "公园", "湖", "森林"}) {
		keywords = append(keywords, "公园", "自然风景")
	}

	keywords = append(keywords, effectiveSearchKeywords(attractionPreference, avoidMuseum, avoidShopping, avoidMountain)...)
	keywords = append(keywords, effectiveSearchKeywords(foodPreference, avoidMuseum, avoidShopping, avoidMountain)...)

	if needsGenericMainAttractionSearch(attractionPreference, foodPreference) {
		keywords = append([]string{"景点", "城市地标"}, keywords...)
	}

	if len(keywords) == 0 {
		keywords = append(keywords, "景点")
	}

	if avoidMuseum {
		keywords = removeMuseumSearchKeywords(keywords)
	}
	if avoidShopping {
		keywords = removeShoppingSearchKeywords(keywords)
	}
	if avoidMountain {
		keywords = removeMountainSearchKeywords(keywords)
	}

	return limitStrings(uniqueStrings(keywords), 8)
}

func (t PreferenceAdapterTool) BuildFallbackAttractionSearchKeywords(
	request model.TripRequest,
	attractions []model.Attraction,
	constraints []model.PreferenceConstraint,
) []string {
	groups := t.BuildFallbackAttractionSearchGroups(request, attractions, constraints)
	keywords := make([]string, 0)

	for _, group := range groups {
		if len(group.Keywords) == 0 {
			continue
		}
		keywords = append(keywords, group.Keywords[0])
	}

	for _, group := range groups {
		if len(group.Keywords) <= 1 {
			continue
		}
		keywords = append(keywords, group.Keywords[1:]...)
	}

	return limitStrings(uniqueStrings(keywords), 8)
}

func (t PreferenceAdapterTool) BuildFallbackAttractionSearchGroups(
	request model.TripRequest,
	attractions []model.Attraction,
	constraints []model.PreferenceConstraint,
) []model.POIFallbackSearchGroup {
	_ = t
	profile := ResolvePreferenceConstraints(constraints)
	groups := make([]model.POIFallbackSearchGroup, 0)

	if requiresSeaIntent(request, profile) && !attractionsContainAnyTag(attractions, []string{"sea", "beach", "waterfront", "coast"}) {
		groups = append(groups, model.POIFallbackSearchGroup{
			Intent:   "sea",
			Keywords: []string{"海边", "海滩", "沙滩", "海滨", "滨海步道"},
		})
	}

	if requiresFoodIntent(profile) && !attractionsContainAnyTag(attractions, []string{"food", "local_food", "snack_street"}) {
		groups = append(groups, model.POIFallbackSearchGroup{
			Intent:   "local_food",
			Keywords: []string{"小吃街", "美食街", "夜市", "老字号"},
		})
	}

	if requiresOldStreetIntent(profile) && !attractionsContainAnyTag(attractions, []string{"old_street", "city_walk", "historic_site"}) {
		groups = append(groups, model.POIFallbackSearchGroup{
			Intent:   "old_street",
			Keywords: []string{"老街", "古街", "历史街区", "文化街区", "步行街"},
		})
	}

	if requiresNightViewIntent(profile) && !attractionsContainAnyTag(attractions, []string{"night_view", "landmark", "waterfront"}) {
		groups = append(groups, model.POIFallbackSearchGroup{
			Intent:   "night_view",
			Keywords: []string{"夜景", "观景", "城市地标", "江景", "海滨夜景"},
		})
	}

	if requiresFamilyScienceIntent(profile) && !attractionsContainStrictFamilyScience(attractions) {
		groups = append(groups, model.POIFallbackSearchGroup{
			Intent:   "family_science",
			Keywords: []string{"科技馆", "科学中心", "天文馆", "自然博物馆", "儿童科技馆", "海洋馆"},
		})
	}

	if hasEffectiveHardAvoidTag(profile.Attraction, "museum") ||
		hasEffectiveHardAvoidTag(profile.Attraction, "exhibition") ||
		hasEffectiveHardAvoidTag(profile.Attraction, "art_gallery") {
		groups = filterFallbackSearchGroups(groups, removeMuseumSearchKeywords)
	}

	if hasEffectiveHardAvoidTag(profile.Attraction, "shopping") ||
		hasEffectiveHardAvoidTag(profile.Attraction, "commercial_area") {
		groups = filterFallbackSearchGroups(groups, removeShoppingSearchKeywords)
	}

	if hasEffectiveHardAvoidTag(profile.Attraction, "mountain") ||
		hasEffectiveHardAvoidTag(profile.Attraction, "hiking") ||
		hasEffectiveHardAvoidTag(profile.Attraction, "high_exertion") {
		groups = filterFallbackSearchGroups(groups, removeMountainSearchKeywords)
	}

	result := make([]model.POIFallbackSearchGroup, 0, len(groups))
	for _, group := range groups {
		group.Keywords = uniqueStrings(group.Keywords)
		if len(group.Keywords) == 0 {
			continue
		}

		result = append(result, group)
	}

	return result
}

func filterFallbackSearchGroups(
	groups []model.POIFallbackSearchGroup,
	filter func([]string) []string,
) []model.POIFallbackSearchGroup {
	result := make([]model.POIFallbackSearchGroup, 0, len(groups))
	for _, group := range groups {
		group.Keywords = filter(group.Keywords)
		result = append(result, group)
	}

	return result
}

func (t PreferenceAdapterTool) FilterAttractions(
	attractions []model.Attraction,
	constraints []model.PreferenceConstraint,
) []model.Attraction {
	if len(attractions) == 0 {
		return attractions
	}

	profiles := make([]model.POIProfile, 0)

	for _, attraction := range attractions {
		profile := BuildPOIProfile(attraction)

		if profile.Invalid {
			continue
		}

		if violatesHardTagConstraint(profile, constraints, "attraction") {
			continue
		}

		profile.Score = scorePOIProfile(profile, constraints)
		profiles = append(profiles, profile)
	}

	if len(profiles) == 0 {
		for _, attraction := range attractions {
			profile := BuildPOIProfile(attraction)
			if !profile.Invalid {
				profile.Score = 0
				profiles = append(profiles, profile)
			}
		}
	}

	sort.SliceStable(profiles, func(i, j int) bool {
		if profiles[i].Score != profiles[j].Score {
			return profiles[i].Score > profiles[j].Score
		}

		return profiles[i].Name < profiles[j].Name
	})

	result := make([]model.Attraction, 0, len(profiles))
	for _, profile := range profiles {
		result = append(result, profile.Attraction)
	}

	return result
}

func violatesHardTagConstraint(
	profile model.POIProfile,
	constraints []model.PreferenceConstraint,
	domain string,
) bool {
	profileText := strings.ToLower(strings.Join([]string{
		profile.Name,
		profile.Category,
		profile.Description,
		profile.Address,
		strings.Join(profile.Tags, " "),
	}, " "))

	for _, constraint := range constraints {
		if !constraintAppliesToDomainInAdapter(constraint, domain) {
			continue
		}

		if strings.ToLower(constraint.Strength) != "hard" {
			continue
		}

		for _, tag := range constraint.AvoidTags {
			if hasTag(profile.Tags, tag) {
				return true
			}
		}

		for _, keyword := range constraint.ExcludeKeywords {
			if keyword != "" && strings.Contains(profileText, strings.ToLower(keyword)) {
				return true
			}
		}
	}

	return false
}

func scorePOIProfile(profile model.POIProfile, constraints []model.PreferenceConstraint) int {
	score := 0

	if hasAnyTag(profile.Tags, []string{"museum", "exhibition", "art_gallery", "memorial", "science_museum"}) {
		score += 20
	}

	if hasAnyTag(profile.Tags, []string{"culture", "indoor", "low_exertion"}) {
		score += 8
	}

	if hasAnyTag(profile.Tags, []string{"high_exertion", "mountain", "hiking"}) {
		score -= 15
	}

	if hasAnyTag(profile.Tags, []string{"food", "shopping"}) {
		score -= 5
	}

	for _, constraint := range constraints {
		if !constraintAppliesToPOIProfile(constraint, profile) {
			continue
		}

		weight := constraintWeight(constraint)

		for _, tag := range constraint.PreferTags {
			if hasTag(profile.Tags, tag) {
				score += weight
			}
		}

		for _, tag := range constraint.AvoidTags {
			if hasTag(profile.Tags, tag) {
				if strings.ToLower(constraint.Strength) == "hard" {
					score -= 1000
				} else {
					score -= weight
				}
			}
		}

		text := strings.ToLower(strings.Join([]string{
			profile.Name,
			profile.Category,
			profile.Description,
			profile.Address,
		}, " "))

		for _, keyword := range constraint.PreferKeywords {
			if keyword != "" && strings.Contains(text, strings.ToLower(keyword)) {
				score += weight / 2
			}
		}

		for _, keyword := range constraint.ExcludeKeywords {
			if keyword != "" && strings.Contains(text, strings.ToLower(keyword)) {
				if strings.ToLower(constraint.Strength) == "hard" {
					score -= 1000
				} else {
					score -= weight
				}
			}
		}
	}

	return score
}

func constraintWeight(constraint model.PreferenceConstraint) int {
	priority := constraint.Priority
	if priority <= 0 {
		priority = 50
	}

	if strings.ToLower(constraint.Source) == "current_request" {
		priority += 50
	}

	if strings.ToLower(constraint.Strength) == "hard" {
		priority += 30
	}

	if priority > 100 {
		priority = 100
	}

	if priority < 10 {
		priority = 10
	}

	return priority / 5
}

func hasTag(tags []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))

	for _, tag := range tags {
		if strings.ToLower(strings.TrimSpace(tag)) == target {
			return true
		}
	}

	return false
}

func hasAnyTag(tags []string, targets []string) bool {
	for _, target := range targets {
		if hasTag(tags, target) {
			return true
		}
	}

	return false
}

func constraintsForPOISearch(constraints []model.PreferenceConstraint) []model.PreferenceConstraint {
	current := make([]model.PreferenceConstraint, 0)

	for _, constraint := range constraints {
		if strings.ToLower(strings.TrimSpace(constraint.Source)) != "current_request" {
			continue
		}

		domain := strings.ToLower(strings.TrimSpace(constraint.Domain))
		if domain != "attraction" && domain != "food" && domain != "route" && domain != "general" {
			continue
		}

		current = append(current, constraint)
	}

	return current
}

func effectiveSearchKeywords(
	preference model.EffectiveDomainPreference,
	avoidMuseum bool,
	avoidShopping bool,
	avoidMountain bool,
) []string {
	keywords := make([]string, 0)

	preferKeywords := append([]string{}, preference.HardPreferKeywords...)
	preferKeywords = append(preferKeywords, preference.SoftPreferKeywords...)
	if avoidMuseum {
		preferKeywords = removeMuseumSearchKeywords(preferKeywords)
	}
	if avoidShopping {
		preferKeywords = removeShoppingSearchKeywords(preferKeywords)
	}
	if avoidMountain {
		preferKeywords = removeMountainSearchKeywords(preferKeywords)
	}
	keywords = append(keywords, preferKeywords...)

	preferTags := append([]string{}, preference.HardPreferTags...)
	preferTags = append(preferTags, preference.SoftPreferTags...)
	for _, tag := range preferTags {
		if avoidMuseum && isMuseumSearchTag(tag) {
			continue
		}

		if avoidShopping && isShoppingSearchTag(tag) {
			continue
		}

		if avoidMountain && isMountainSearchTag(tag) {
			continue
		}

		if hasEffectiveHardAvoidTag(preference, tag) {
			continue
		}

		keywords = append(keywords, keywordsForPOITag(tag)...)
	}

	return keywords
}

func needsGenericMainAttractionSearch(
	attractionPreference model.EffectiveDomainPreference,
	foodPreference model.EffectiveDomainPreference,
) bool {
	hasFoodIntent := hasString(foodPreference.HardPreferTags, "food") ||
		hasString(foodPreference.SoftPreferTags, "food") ||
		hasString(foodPreference.HardPreferTags, "local_food") ||
		hasString(foodPreference.SoftPreferTags, "local_food") ||
		hasString(foodPreference.HardPreferTags, "snack_street") ||
		hasString(foodPreference.SoftPreferTags, "snack_street")

	if !hasFoodIntent {
		return false
	}

	mainTags := []string{
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
	}

	for _, tag := range mainTags {
		if hasString(attractionPreference.HardPreferTags, tag) || hasString(attractionPreference.SoftPreferTags, tag) {
			return false
		}
	}

	return true
}

func hasEffectiveHardAvoidTag(preference model.EffectiveDomainPreference, target string) bool {
	return hasString(preference.HardAvoidTags, target)
}

func requiresSeaIntent(request model.TripRequest, profile model.EffectivePreferenceProfile) bool {
	text := strings.ToLower(request.RawInput + " " + request.Preference)

	if containsAnyText(text, []string{"看海", "海边", "海滩", "沙滩", "海滨", "海岸", "滨海", "观海"}) {
		return true
	}

	for _, tag := range []string{"sea", "beach", "waterfront", "coast"} {
		if hasString(profile.Attraction.HardPreferTags, tag) || hasString(profile.Attraction.SoftPreferTags, tag) {
			return true
		}
	}

	return false
}

func requiresFoodIntent(profile model.EffectivePreferenceProfile) bool {
	return effectiveDomainContainsAnyTag(profile.Food, []string{"food", "local_food", "snack_street", "night_market"}) ||
		effectiveDomainContainsAnyTag(profile.Attraction, []string{"food", "local_food", "snack_street", "night_market"})
}

func requiresOldStreetIntent(profile model.EffectivePreferenceProfile) bool {
	return effectiveDomainContainsAnyTag(profile.Attraction, []string{"old_street", "city_walk", "historic_site"})
}

func requiresNightViewIntent(profile model.EffectivePreferenceProfile) bool {
	return effectiveDomainContainsAnyTag(profile.Attraction, []string{"night_view", "landmark"})
}

func requiresFamilyScienceIntent(profile model.EffectivePreferenceProfile) bool {
	return effectiveDomainContainsAnyTag(profile.Attraction, []string{
		"family",
		"science_museum",
		"astronomy",
		"aquarium",
		"children_exhibition",
		"natural_science",
		"child_friendly",
	})
}

func effectiveDomainContainsAnyTag(preference model.EffectiveDomainPreference, tags []string) bool {
	for _, tag := range tags {
		if hasString(preference.HardPreferTags, tag) || hasString(preference.SoftPreferTags, tag) {
			return true
		}
	}

	return false
}

func attractionsContainAnyTag(attractions []model.Attraction, tags []string) bool {
	for _, attraction := range attractions {
		profile := BuildPOIProfile(attraction)
		if hasAnyTag(profile.Tags, tags) {
			return true
		}
	}

	return false
}

func attractionsContainStrictFamilyScience(attractions []model.Attraction) bool {
	for _, attraction := range attractions {
		poi := BuildTripPOI(attraction, 0)
		if poi.Role != tripPOIRoleMainAttraction {
			continue
		}

		if hasAnyTag(poi.Tags, []string{"science_museum", "astronomy", "aquarium", "children_exhibition", "natural_science"}) {
			return true
		}
	}

	return false
}

func hasCurrentHardAvoidTag(constraints []model.PreferenceConstraint, domain string, target string) bool {
	for _, constraint := range constraints {
		if strings.ToLower(strings.TrimSpace(constraint.Source)) != "current_request" {
			continue
		}

		if strings.ToLower(strings.TrimSpace(constraint.Strength)) != "hard" {
			continue
		}

		if !constraintAppliesToDomainInAdapter(constraint, domain) {
			continue
		}

		if hasTag(constraint.AvoidTags, target) {
			return true
		}
	}

	return false
}

func isMuseumSearchTag(tag string) bool {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "museum", "exhibition", "art_gallery":
		return true
	default:
		return false
	}
}

func isShoppingSearchTag(tag string) bool {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "shopping", "commercial_area":
		return true
	default:
		return false
	}
}

func isMountainSearchTag(tag string) bool {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "mountain", "hiking", "high_exertion":
		return true
	default:
		return false
	}
}

func removeMuseumSearchKeywords(keywords []string) []string {
	result := make([]string, 0, len(keywords))

	for _, keyword := range keywords {
		if containsAnyText(keyword, []string{"博物馆", "展览馆", "展览中心", "美术馆", "艺术馆"}) {
			continue
		}

		result = append(result, keyword)
	}

	return result
}

func removeShoppingSearchKeywords(keywords []string) []string {
	result := make([]string, 0, len(keywords))

	for _, keyword := range keywords {
		if containsAnyText(keyword, []string{"商场", "购物中心", "商业街", "市场", "百货"}) {
			continue
		}

		result = append(result, keyword)
	}

	return result
}

func removeMountainSearchKeywords(keywords []string) []string {
	result := make([]string, 0, len(keywords))

	for _, keyword := range keywords {
		if containsAnyText(keyword, []string{"山", "登山", "爬山", "徒步", "峡谷"}) {
			continue
		}

		result = append(result, keyword)
	}

	return result
}

func constraintCanGuidePOISearch(constraint model.PreferenceConstraint) bool {
	return constraintAppliesToDomainInAdapter(constraint, "attraction") ||
		strings.ToLower(strings.TrimSpace(constraint.Domain)) == "food"
}

func constraintAppliesToPOIProfile(constraint model.PreferenceConstraint, profile model.POIProfile) bool {
	if constraintAppliesToDomainInAdapter(constraint, "attraction") {
		return true
	}

	if strings.ToLower(strings.TrimSpace(constraint.Domain)) == "food" {
		return hasAnyTag(profile.Tags, []string{"food", "local_food"})
	}

	return false
}

func keywordsForPOITag(tag string) []string {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "museum":
		return []string{"博物馆"}
	case "exhibition":
		return []string{"展览馆", "展览中心"}
	case "art_gallery":
		return []string{"美术馆", "艺术馆"}
	case "memorial":
		return []string{"纪念馆"}
	case "culture":
		return []string{"历史", "人文"}
	case "indoor":
		return []string{"室内景点", "科技馆", "海洋馆"}
	case "low_exertion":
		return []string{"室内景点", "城市地标"}
	case "old_street", "historic_site", "city_walk":
		return []string{"老街", "古街", "历史街区", "步行街", "市区漫步", "文化街区"}
	case "science_museum":
		return []string{"科技馆", "科学中心"}
	case "family":
		return []string{"亲子", "科技馆", "海洋馆"}
	case "zoo":
		return []string{"动物园"}
	case "aquarium":
		return []string{"海洋馆", "水族馆"}
	case "shopping", "commercial_area":
		return []string{"商场", "购物中心", "商业街"}
	case "food", "local_food", "snack_street":
		return []string{"小吃街", "美食街", "本地小吃", "夜市", "老字号"}
	case "night_view":
		return []string{"夜景", "观景", "江景", "海滨夜景", "城市夜景"}
	case "landmark":
		return []string{"地标", "观景"}
	case "sea", "beach", "waterfront", "coast":
		return []string{"海边", "海滩", "沙滩", "海滨", "海岸", "滨海步道", "观海"}
	case "nature", "park":
		return []string{"公园", "自然风景"}
	case "lake":
		return []string{"湖"}
	case "forest":
		return []string{"森林公园"}
	default:
		return nil
	}
}

func uniqueStrings(items []string) []string {
	result := make([]string, 0, len(items))
	seen := map[string]bool{}

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		key := strings.ToLower(item)
		if seen[key] {
			continue
		}

		seen[key] = true
		result = append(result, item)
	}

	return result
}

func limitStrings(items []string, limit int) []string {
	if limit <= 0 || len(items) <= limit {
		return items
	}

	return items[:limit]
}

func buildAttractionConstraintText(attraction model.Attraction) string {
	return strings.ToLower(strings.Join([]string{
		attraction.Name,
		attraction.Category,
		attraction.Address,
		attraction.Description,
		attraction.VisitTime,
		attraction.DataSource,
	}, " "))
}

func (t PreferenceAdapterTool) AdjustDailyRoutes(
	routes []model.DailyRoute,
	constraints []model.PreferenceConstraint,
) []model.DailyRoute {
	if len(routes) == 0 {
		return routes
	}

	result := make([]model.DailyRoute, len(routes))
	copy(result, routes)

	if prefersRelaxedRoute(constraints) {
		for i := range result {
			result[i] = keepFirstNRouteAttractions(result[i], 2)
			result[i].Summary = appendPreferenceSummary(result[i].Summary, "已根据用户轻松旅行偏好控制单日景点数量。")
			result[i].DataSource = appendPreferenceDataSource(result[i].DataSource, "preference_route_relaxed")
		}
	}

	if avoidsLongTransfer(constraints) {
		for i := range result {
			result[i] = removeRouteAfterLongTransfer(result[i], 75)
			result[i].Summary = appendPreferenceSummary(result[i].Summary, "已根据用户不想长距离奔波的偏好减少长通勤路线。")
			result[i].DataSource = appendPreferenceDataSource(result[i].DataSource, "preference_route_short_transfer")
		}
	}

	if prefersIntensiveRoute(constraints) {
		for i := range result {
			result[i].Summary = appendPreferenceSummary(result[i].Summary, "用户偏好紧凑行程，已尽量保留更多可行景点。")
			result[i].DataSource = appendPreferenceDataSource(result[i].DataSource, "preference_route_intensive")
		}
	}

	return result
}

func keepFirstNRouteAttractions(route model.DailyRoute, n int) model.DailyRoute {
	if n < 0 {
		n = 0
	}

	if len(route.Attractions) > n {
		route.Attractions = route.Attractions[:n]
	}

	if n <= 1 {
		route.RouteSegments = route.RouteSegments[:0]
		return route
	}

	maxSegments := n - 1
	if len(route.RouteSegments) > maxSegments {
		route.RouteSegments = route.RouteSegments[:maxSegments]
	}

	return route
}

func removeRouteAfterLongTransfer(route model.DailyRoute, maxMinutes int) model.DailyRoute {
	if len(route.RouteSegments) == 0 {
		return route
	}

	cutAttractionCount := len(route.Attractions)

	for i, segment := range route.RouteSegments {
		if segment.DurationMinutes > maxMinutes {
			cutAttractionCount = i + 1
			break
		}
	}

	return keepFirstNRouteAttractions(route, cutAttractionCount)
}

func prefersRelaxedRoute(constraints []model.PreferenceConstraint) bool {
	return routeConstraintContains(constraints, []string{
		"relaxed",
		"slow_pace",
		"short_transfer",
	})
}

func prefersIntensiveRoute(constraints []model.PreferenceConstraint) bool {
	return routeConstraintContains(constraints, []string{
		"intensive",
		"more_attractions",
		"tight_schedule",
	})
}

func avoidsLongTransfer(constraints []model.PreferenceConstraint) bool {
	return routeConstraintContains(constraints, []string{
		"long_transfer",
		"short_transfer",
	})
}

func routeConstraintContains(constraints []model.PreferenceConstraint, keywords []string) bool {
	for _, constraint := range constraints {
		if !constraintAppliesToDomainInAdapter(constraint, "route") {
			continue
		}

		parts := make([]string, 0)
		parts = append(parts, constraint.PreferKeywords...)
		parts = append(parts, constraint.ExcludeKeywords...)
		parts = append(parts, constraint.PreferTags...)
		parts = append(parts, constraint.AvoidTags...)
		text := strings.ToLower(strings.Join(parts, " "))

		for _, keyword := range keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				return true
			}
		}
	}

	return false
}

func appendPreferenceSummary(summary string, addition string) string {
	if summary == "" {
		return addition
	}

	if strings.Contains(summary, addition) {
		return summary
	}

	return summary + addition
}

func appendPreferenceDataSource(dataSource string, suffix string) string {
	if dataSource == "" {
		return suffix
	}

	if strings.Contains(dataSource, suffix) {
		return dataSource
	}

	return dataSource + "_" + suffix
}

func violatesHardConstraint(
	targetText string,
	constraints []model.PreferenceConstraint,
	domain string,
) bool {
	targetText = strings.ToLower(targetText)

	for _, constraint := range constraints {
		if !constraintAppliesToDomainInAdapter(constraint, domain) {
			continue
		}

		if strings.ToLower(constraint.Strength) != "hard" {
			continue
		}

		for _, keyword := range constraint.ExcludeKeywords {
			if keyword == "" {
				continue
			}

			if strings.Contains(targetText, strings.ToLower(keyword)) {
				return true
			}
		}
	}

	return false
}

func scoreTextByConstraints(
	targetText string,
	constraints []model.PreferenceConstraint,
	domain string,
) int {
	targetText = strings.ToLower(targetText)

	score := 0

	for _, constraint := range constraints {
		if !constraintAppliesToDomainInAdapter(constraint, domain) {
			continue
		}

		weight := constraintWeight(constraint)

		for _, keyword := range constraint.PreferKeywords {
			if keyword != "" && strings.Contains(targetText, strings.ToLower(keyword)) {
				score += weight / 2
			}
		}

		for _, keyword := range constraint.ExcludeKeywords {
			if keyword == "" || !strings.Contains(targetText, strings.ToLower(keyword)) {
				continue
			}

			if strings.ToLower(constraint.Strength) == "hard" {
				score -= 100
			} else {
				score -= weight
			}
		}
	}

	return score
}

func constraintAppliesToDomainInAdapter(constraint model.PreferenceConstraint, domain string) bool {
	constraintDomain := strings.ToLower(strings.TrimSpace(constraint.Domain))

	return constraintDomain == "" ||
		constraintDomain == "general" ||
		constraintDomain == strings.ToLower(domain)
}
