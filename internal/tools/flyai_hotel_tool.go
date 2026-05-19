package tools

import (
	"context"
	"strings"

	"ariadne/internal/model"
	"ariadne/internal/provider/flyai"
)

type FlyAIHotelTool struct {
	Name        string
	Description string
	Client      flyai.Client
}

func NewFlyAIHotelTool() FlyAIHotelTool {
	return FlyAIHotelTool{
		Name:        "flyai_hotel_tool",
		Description: "使用 FlyAI / 飞猪真实酒店商品数据搜索酒店报价",
		Client:      flyai.NewClient(),
	}
}

func (t FlyAIHotelTool) Run(
	request model.TripRequest,
	budgetBreakdown model.BudgetBreakdown,
	attractions []model.Attraction,
) []model.HotelOffer {
	return t.RunWithPreferences(request, budgetBreakdown, attractions, nil)
}

func (t FlyAIHotelTool) RunWithPreferences(
	request model.TripRequest,
	budgetBreakdown model.BudgetBreakdown,
	attractions []model.Attraction,
	constraints []model.PreferenceConstraint,
) []model.HotelOffer {
	if request.StartDate == "" || request.EndDate == "" {
		return []model.HotelOffer{
			{
				Provider:   "fliggy",
				DataSource: "flyai_fliggy",
				Status:     "unavailable",
				Message:    "缺少入住日期或离店日期，无法查询真实酒店价格。",
			},
		}
	}

	searchTerms := buildHotelSearchTerms(request, attractions, constraints)

	maxPrice := budgetBreakdown.HotelBudgetPerNight
	if maxPrice <= 0 {
		maxPrice = 500
	}

	if hotelPrefersHighEndFromConstraints(constraints) && maxPrice < 1500 {
		maxPrice = 1500
	}

	if hotelPrefersUniqueStayFromConstraints(constraints) && maxPrice < 900 {
		maxPrice = 900
	}

	allOffers := make([]model.HotelOffer, 0)
	var lastErr error
	for _, term := range searchTerms {
		offers, err := t.Client.SearchHotels(
			context.Background(),
			request.Destination,
			term,
			request.StartDate,
			request.EndDate,
			maxPrice,
		)

		if err != nil {
			lastErr = err
			continue
		}

		allOffers = append(allOffers, offers...)
	}

	allOffers = deduplicateHotelOffers(allOffers)
	if len(allOffers) > 0 {
		return allOffers
	}

	message := "没有找到可用的真实酒店报价。"
	if lastErr != nil {
		message = lastErr.Error()
	}

	if len(searchTerms) > 0 {
		message = "按当前住宿偏好关键词 " + strings.Join(searchTerms, "、") + " 未找到可用的真实酒店报价。"
		if lastErr != nil {
			message += " " + lastErr.Error()
		}
	}

	return []model.HotelOffer{
		{
			Provider:   "fliggy",
			DataSource: "flyai_fliggy",
			Status:     "unavailable",
			Message:    message,
		},
	}
}

func buildHotelSearchTerms(
	request model.TripRequest,
	attractions []model.Attraction,
	constraints []model.PreferenceConstraint,
) []string {
	terms := make([]string, 0)

	if hotelPrefersUniqueStayFromConstraints(constraints) || requestPrefersUniqueStay(request) {
		terms = append(terms, "民宿", "客栈", "特色住宿", "老城", "古城", "老街")
	}

	if hotelPrefersHighEndFromConstraints(constraints) || requestPrefersHighEndHotel(request) {
		terms = append(terms, "高端酒店", "度假酒店", "海景酒店", "海边酒店")
	}

	if len(terms) == 0 {
		terms = append(terms, chooseHotelPOI(attractions))
	}

	return uniqueStrings(limitStrings(terms, 6))
}

func hotelPrefersUniqueStayFromConstraints(constraints []model.PreferenceConstraint) bool {
	for _, constraint := range constraints {
		if strings.ToLower(strings.TrimSpace(constraint.Source)) != "current_request" {
			continue
		}

		if !constraintAppliesToDomainInAdapter(constraint, "hotel") {
			continue
		}

		if hasAnyTag(constraint.PreferTags, []string{"homestay", "guesthouse", "unique_stay"}) {
			return true
		}
	}

	return false
}

func hotelPrefersHighEndFromConstraints(constraints []model.PreferenceConstraint) bool {
	for _, constraint := range constraints {
		if strings.ToLower(strings.TrimSpace(constraint.Source)) != "current_request" {
			continue
		}

		if !constraintAppliesToDomainInAdapter(constraint, "hotel") {
			continue
		}

		if hasAnyTag(constraint.PreferTags, []string{"high_end_hotel", "resort", "sea_nearby", "comfort_hotel"}) {
			return true
		}
	}

	return false
}

func requestPrefersUniqueStay(request model.TripRequest) bool {
	text := strings.ToLower(request.RawInput + " " + request.Preference)
	return containsAnyText(text, []string{"想体验民宿", "想住民宿", "海边民宿", "特色民宿", "客栈", "特色客栈", "特色住宿"})
}

func requestPrefersHighEndHotel(request model.TripRequest) bool {
	text := strings.ToLower(request.RawInput + " " + request.Preference)
	return containsAnyText(text, []string{"高端酒店", "度假酒店", "住好一点", "海边酒店", "海景酒店", "resort", "四星以上", "五星"})
}

func deduplicateHotelOffers(offers []model.HotelOffer) []model.HotelOffer {
	result := make([]model.HotelOffer, 0, len(offers))
	seen := map[string]bool{}

	for _, offer := range offers {
		key := strings.ToLower(strings.TrimSpace(offer.Name) + "|" + strings.TrimSpace(offer.Address))
		if key == "|" || seen[key] {
			continue
		}

		seen[key] = true
		result = append(result, offer)
	}

	return result
}

func legacyHotelUnavailable(err error) []model.HotelOffer {
	if err != nil {
		return []model.HotelOffer{
			{
				Provider:   "fliggy",
				DataSource: "flyai_fliggy",
				Status:     "unavailable",
				Message:    err.Error(),
			},
		}
	}

	return nil
}

func chooseHotelPOI(attractions []model.Attraction) string {
	if len(attractions) == 0 {
		return ""
	}

	for _, attraction := range attractions {
		if attraction.Name != "" {
			return attraction.Name
		}
	}

	return ""
}
