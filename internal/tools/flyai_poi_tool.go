package tools

import (
	"context"

	"ariadne/internal/model"
	"ariadne/internal/provider/flyai"
)

type FlyAIPoiTool struct {
	Name        string
	Description string
	Client      flyai.Client
}

func NewFlyAIPoiTool() FlyAIPoiTool {
	return FlyAIPoiTool{
		Name:        "flyai_poi_tool",
		Description: "使用 FlyAI / 飞猪搜索景点详情、图片和跳转链接",
		Client:      flyai.NewClient(),
	}
}

func (t FlyAIPoiTool) Run(request model.TripRequest, attractions []model.Attraction) []model.PoiOffer {
	if request.Destination == "" {
		return []model.PoiOffer{
			{
				Provider:   "fliggy",
				DataSource: "flyai_fliggy",
				Status:     "unavailable",
				Message:    "缺少目的地城市，无法查询 FlyAI 景点信息。",
			},
		}
	}

	if len(attractions) == 0 {
		return []model.PoiOffer{
			{
				Provider:   "fliggy",
				DataSource: "flyai_fliggy",
				Status:     "unavailable",
				Message:    "没有可用于查询的景点列表。",
			},
		}
	}

	result := make([]model.PoiOffer, 0)
	seen := make(map[string]bool)

	limit := len(attractions)
	if limit > 5 {
		limit = 5
	}

	for i := 0; i < limit; i++ {
		keyword := attractions[i].Name
		if keyword == "" {
			continue
		}

		offers, err := t.Client.SearchPOIs(context.Background(), request.Destination, keyword)
		if err != nil {
			result = append(result, model.PoiOffer{
				Provider:   "fliggy",
				Name:       keyword,
				DataSource: "flyai_fliggy",
				Status:     "unavailable",
				Message:    err.Error(),
			})
			continue
		}

		for _, offer := range offers {
			if offer.Name == "" || seen[offer.Name] {
				continue
			}

			seen[offer.Name] = true
			result = append(result, offer)
			break
		}
	}

	return result
}