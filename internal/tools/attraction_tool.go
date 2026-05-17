package tools

import "ariadne/internal/model"

func RecommendAttractions(request model.TripRequest) []model.Attraction {
	switch request.Destination {
	case "杭州":
		return []model.Attraction{
			{
				Name:          "西湖",
				Category:      "自然风景",
				Description:   "杭州最经典的城市湖泊景区，适合散步、骑行和拍照。",
				EstimatedCost: 0,
				VisitTime:     "半天",
				Link:          "https://ditu.amap.com/search?query=西湖",
			},
			{
				Name:          "灵隐寺",
				Category:      "历史人文",
				Description:   "杭州著名寺院，适合喜欢历史文化和安静氛围的游客。",
				EstimatedCost: 75,
				VisitTime:     "半天",
				Link:          "https://ditu.amap.com/search?query=灵隐寺",
			},
			{
				Name:          "河坊街",
				Category:      "城市街区",
				Description:   "适合晚上逛吃，能体验杭州老街氛围和小吃。",
				EstimatedCost: 100,
				VisitTime:     "2-3小时",
				Link:          "https://ditu.amap.com/search?query=河坊街",
			},
		}
	default:
		return []model.Attraction{
			{
				Name:          request.Destination + "市中心",
				Category:      "城市探索",
				Description:   "先从目的地市中心开始，适合初次到达后熟悉城市环境。",
				EstimatedCost: 0,
				VisitTime:     "2-3小时",
				Link:          "https://ditu.amap.com/search?query=" + request.Destination,
			},
		}
	}
}