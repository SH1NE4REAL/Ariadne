package tools

import "ariadne/internal/model"

type AttractionTool struct {
	Name        string
	Description string
}

func NewAttractionTool() AttractionTool {
	return AttractionTool{
		Name:        "attraction_tool",
		Description: "根据目的地、旅行天数和用户偏好推荐景点",
	}
}

func (t AttractionTool) Run(request model.TripRequest) []model.Attraction {
	return RecommendAttractions(request)
}