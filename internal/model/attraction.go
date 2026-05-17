package model

type Attraction struct {
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	Address       string  `json:"address"`
	Description   string  `json:"description"`
	EstimatedCost int     `json:"estimated_cost"`
	VisitTime     string  `json:"visit_time"`
	Link          string  `json:"link"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	DataSource    string  `json:"data_source"`
}