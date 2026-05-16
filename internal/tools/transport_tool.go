package tools

import "ariadne/internal/model"

func GenerateTransportPlans(request model.TripRequest) []model.TransportPlan {
	plans := []model.TransportPlan{
		{
			Method:      "高铁",
			Duration:    "约1.5-3小时",
			Price:       200,
			Description: "速度快、稳定性高，适合中短途城市旅行。",
			BookingLink: "https://www.12306.cn/index/",
		},
		{
			Method:      "飞机",
			Duration:    "约2-4小时，含机场通勤和安检时间",
			Price:       600,
			Description: "适合远距离出行，但市区到机场的时间成本较高。",
			BookingLink: "https://flights.ctrip.com/",
		},
		{
			Method:      "自驾",
			Duration:    "根据距离变化较大",
			Price:       400,
			Description: "自由度高，适合多人出行或沿途游玩。",
			BookingLink: "https://ditu.amap.com/",
		},
	}

	return plans
}