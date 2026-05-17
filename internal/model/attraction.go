package model

type Attraction struct {
	Name          string `json:"name"`
	Category      string `json:"category"`
	Description   string `json:"description"`
	EstimatedCost int    `json:"estimated_cost"`
	VisitTime     string `json:"visit_time"`
	Link          string `json:"link"`
}