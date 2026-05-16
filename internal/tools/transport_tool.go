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

func RecommendTransportPlan(request model.TripRequest, plans []model.TransportPlan) model.TransportPlan {
	if len(plans) == 0 {
		return model.TransportPlan{}
	}

	if request.Preference == "轻松" {
		for _, plan := range plans {
			if plan.Method == "高铁" {
				plan.Reason = "高铁时间稳定、乘坐舒适，省去了机场通勤和安检等待，适合轻松出行。"
				return plan
			}
		}
	}

	if request.Preference == "省钱" {
		for _, plan := range plans {
			if plan.Method == "高铁" {
				plan.Reason = "高铁价格通常低于飞机，同时速度和舒适度都比较均衡。"
				return plan
			}
		}
	}

	for _, plan := range plans {
		if plan.Price <= request.Budget/3 {
			plan.Reason = "该方案价格占总预算比例较低，适合当前预算。"
			return plan
		}
	}

	plans[0].Reason = "默认推荐该方案，后续会结合真实价格和路线数据进一步优化。"
	return plans[0]
}