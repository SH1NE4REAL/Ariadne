package tools

import (
	"context"

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

	poiName := chooseHotelPOI(attractions)

	maxPrice := budgetBreakdown.HotelBudgetPerNight
	if maxPrice <= 0 {
		maxPrice = 500
	}

	offers, err := t.Client.SearchHotels(
		context.Background(),
		request.Destination,
		poiName,
		request.StartDate,
		request.EndDate,
		maxPrice,
	)

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

	return offers
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