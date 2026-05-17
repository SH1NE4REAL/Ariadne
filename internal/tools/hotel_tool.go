package tools

import (
	"net/url"

	"ariadne/internal/model"
)

type HotelTool struct {
	Name        string
	Description string
}

func NewHotelTool() HotelTool {
	return HotelTool{
		Name:        "hotel_tool",
		Description: "根据目的地、旅行天数和每晚住宿预算推荐住宿档位",
	}
}

func (t HotelTool) Run(request model.TripRequest, budgetBreakdown model.BudgetBreakdown) []model.HotelOption {
	return GenerateHotelOptions(request, budgetBreakdown)
}

func GenerateHotelOptions(request model.TripRequest, budgetBreakdown model.BudgetBreakdown) []model.HotelOption {
	nights := request.Days - 1
	if nights <= 0 {
		nights = 1
	}

	pricePerNight := budgetBreakdown.HotelBudgetPerNight
	if pricePerNight <= 0 {
		pricePerNight = 200
	}

	destination := url.QueryEscape(request.Destination)
	bookingLink := "https://hotels.ctrip.com/hotels/list?cityName=" + destination

	if pricePerNight < 150 {
		return []model.HotelOption{
			{
				Name:          request.Destination + "青旅 / 低价民宿",
				Level:         "低预算",
				Location:      "建议选择地铁沿线或非核心景区区域",
				Description:   "适合预算非常有限的旅行，优先考虑床位房、青旅或低价民宿。",
				PricePerNight: pricePerNight,
				Nights:        nights,
				TotalPrice:    pricePerNight * nights,
				BookingLink:   bookingLink,
				Reason:        "当前每晚住宿预算较低，建议优先保证安全和交通便利，不追求酒店配置。",
			},
		}
	}

	if pricePerNight < 300 {
		return []model.HotelOption{
			{
				Name:          request.Destination + "经济型酒店",
				Level:         "经济型",
				Location:      "建议选择地铁站附近或热门景区外围",
				Description:   "适合控制预算的普通旅行，优先选择连锁经济型酒店。",
				PricePerNight: pricePerNight,
				Nights:        nights,
				TotalPrice:    pricePerNight * nights,
				BookingLink:   bookingLink,
				Reason:        "当前预算适合经济型住宿，兼顾价格和基础舒适度。",
			},
		}
	}

	if pricePerNight < 600 {
		return []model.HotelOption{
			{
				Name:          request.Destination + "舒适型酒店",
				Level:         "舒适型",
				Location:      "建议选择核心景区、商圈或地铁换乘站附近",
				Description:   "适合希望减少通勤、提升休息质量的旅行。",
				PricePerNight: pricePerNight,
				Nights:        nights,
				TotalPrice:    pricePerNight * nights,
				BookingLink:   bookingLink,
				Reason:        "当前住宿预算比较充足，可以优先选择位置更好的酒店，减少通勤消耗。",
			},
		}
	}

	return []model.HotelOption{
		{
			Name:          request.Destination + "高品质酒店",
			Level:         "高品质",
			Location:      "建议选择核心景区、湖景区域或高评分商圈酒店",
			Description:   "适合预算宽松、希望提升旅行体验的用户。",
			PricePerNight: pricePerNight,
			Nights:        nights,
			TotalPrice:    pricePerNight * nights,
			BookingLink:   bookingLink,
			Reason:        "当前住宿预算较宽松，可以优先考虑位置、评分和舒适度。",
		},
	}
}