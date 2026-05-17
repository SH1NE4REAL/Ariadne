package model

type DailyRoute struct {
	Day                  int            `json:"day"`
	Title                string         `json:"title"`
	Attractions          []Attraction   `json:"attractions"`
	RouteSegments        []RouteSegment `json:"route_segments"`
	Summary              string         `json:"summary"`
	EstimatedCost        int            `json:"estimated_cost"`
	DataSource           string         `json:"data_source"`
	Optimized            bool           `json:"optimized"`
	OptimizationStrategy string         `json:"optimization_strategy"`
}