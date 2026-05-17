package model

type RouteSegment struct {
	From            string `json:"from"`
	To              string `json:"to"`
	Mode            string `json:"mode"`
	DistanceMeters  int    `json:"distance_meters"`
	DurationSeconds int    `json:"duration_seconds"`
	DurationMinutes int    `json:"duration_minutes"`
	DataSource      string `json:"data_source"`
	Status          string `json:"status"`
	Message         string `json:"message"`
}