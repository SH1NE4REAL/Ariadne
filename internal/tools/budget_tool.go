package tools

import "ariadne/internal/model"

type BudgetTool struct {
	Name        string
	Description string
}

func NewBudgetTool() BudgetTool {
	return BudgetTool{
		Name:        "budget_tool",
		Description: "根据总预算、天数、偏好和最优交通方案拆分旅行预算",
	}
}

func (t BudgetTool) Run(request model.TripRequest, bestBookingOption model.BestBookingOption) model.BudgetBreakdown {
	return GenerateBudgetBreakdown(request, bestBookingOption)
}

func GenerateBudgetBreakdown(request model.TripRequest, bestBookingOption model.BestBookingOption) model.BudgetBreakdown {
	totalBudget := request.Budget
	estimatedTransportCost := bestBookingOption.Best.Price

	if totalBudget <= 0 {
		return model.BudgetBreakdown{
			Status:      "missing_budget",
			Suggestions: []string{"当前缺少预算信息，无法进行预算拆分。"},
		}
	}

	transportBudget := estimatedTransportCost
	if transportBudget <= 0 {
		transportBudget = totalBudget / 3
	}

	remainingBudget := totalBudget - transportBudget

	if remainingBudget <= 0 {
		return model.BudgetBreakdown{
			TotalBudget:            totalBudget,
			TransportBudget:        transportBudget,
			EstimatedTransportCost: estimatedTransportCost,
			RemainingBudget:        remainingBudget,
			Status:                 "over_budget",
			Suggestions: []string{
				"当前交通成本已经接近或超过总预算，建议优先选择更便宜的交通方式。",
				"可以减少旅行天数，或提高总预算。",
			},
		}
	}

	hotelRatio := 40
	foodRatio := 25
	attractionRatio := 20
	

	if request.Preference == "省钱" {
		hotelRatio = 35
		foodRatio = 30
		attractionRatio = 15
		
	}

	if request.Preference == "轻松" {
		hotelRatio = 45
		foodRatio = 25
		attractionRatio = 15
		
	}

	hotelBudget := remainingBudget * hotelRatio / 100
	foodBudget := remainingBudget * foodRatio / 100
	attractionBudget := remainingBudget * attractionRatio / 100
	reserveBudget := remainingBudget - hotelBudget - foodBudget - attractionBudget

	hotelNights := request.Days - 1
	if hotelNights <= 0 {
		hotelNights = 1
	}

	hotelBudgetPerNight := hotelBudget / hotelNights

	suggestions := generateBudgetSuggestions(request, hotelBudgetPerNight, foodBudget, attractionBudget, reserveBudget)

	return model.BudgetBreakdown{
		TotalBudget:            totalBudget,
		TransportBudget:        transportBudget,
		HotelBudget:            hotelBudget,
		HotelBudgetPerNight:    hotelBudgetPerNight,
		FoodBudget:             foodBudget,
		AttractionBudget:       attractionBudget,
		ReserveBudget:          reserveBudget,
		EstimatedTransportCost: estimatedTransportCost,
		RemainingBudget:        remainingBudget,
		Status:                 "ok",
		Suggestions:            suggestions,
	}
}

func generateBudgetSuggestions(request model.TripRequest, hotelBudgetPerNight int, foodBudget int, attractionBudget int, reserveBudget int) []string {
	suggestions := make([]string, 0)

	if hotelBudgetPerNight < 200 {
		suggestions = append(suggestions, "当前每晚住宿预算偏低，建议优先选择青旅、经济型酒店或远离热门景区的住宿。")
	} else if hotelBudgetPerNight < 400 {
		suggestions = append(suggestions, "当前住宿预算适合选择经济型酒店或普通连锁酒店。")
	} else {
		suggestions = append(suggestions, "当前住宿预算相对宽松，可以考虑位置更好的酒店。")
	}

	if request.Preference == "省钱" {
		suggestions = append(suggestions, "你选择了省钱偏好，建议优先控制住宿和交通成本，把预算留给必要餐饮和少量核心景点。")
	}

	if request.Preference == "轻松" {
		suggestions = append(suggestions, "你选择了轻松偏好，建议住宿尽量靠近核心景区或地铁站，减少通勤消耗。")
	}

	if attractionBudget < 100 {
		suggestions = append(suggestions, "景点预算较低，建议优先选择免费景区、公园、街区和城市散步路线。")
	}

	if reserveBudget < 100 {
		suggestions = append(suggestions, "机动预算较少，建议预留一部分给临时交通、行李寄存或突发支出。")
	}

	return suggestions
}