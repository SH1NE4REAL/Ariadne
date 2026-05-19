package tools

import "ariadne/internal/model"

func RecommendAttractions(request model.TripRequest) []model.Attraction {
	return []model.Attraction{
		{
			Name:          request.Destination + "市中心",
			Category:      "城市探索",
			Description:   "先从目的地市中心开始，适合初次到达后熟悉城市环境。",
			EstimatedCost: 0,
			VisitTime:     "2-3小时",
			Link:          "https://ditu.amap.com/search?query=" + request.Destination,
		},
		{
			Name:          request.Destination + "城市街区",
			Category:      "城市街区",
			Description:   "优先选择交通便利、适合步行浏览的城市街区作为通用候选。",
			EstimatedCost: 0,
			VisitTime:     "2-3小时",
			Link:          "https://ditu.amap.com/search?query=" + request.Destination + " 城市街区",
		},
		{
			Name:          request.Destination + "自然风景",
			Category:      "自然风景",
			Description:   "在真实 POI 数据不足时，用目的地的通用自然风景候选补位。",
			EstimatedCost: 0,
			VisitTime:     "2-3小时",
			Link:          "https://ditu.amap.com/search?query=" + request.Destination + " 自然风景",
		},
	}
}
