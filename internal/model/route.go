package model

type DailyRoute struct {
	Day           int          `json:"day"`
	Title         string       `json:"title"`
	Attractions   []Attraction `json:"attractions"`
	Summary       string       `json:"summary"`
	EstimatedCost int          `json:"estimated_cost"`
}