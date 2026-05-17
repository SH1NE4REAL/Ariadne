package model

type RouteDistance struct {
	Origin            string `json:"origin"`
	Destination       string `json:"destination"`
	Mode              string `json:"mode"`
	DistanceMeters    int    `json:"distance_meters"`
	DurationMinutes   int    `json:"duration_minutes"`
	EstimatedTaxiFare int    `json:"estimated_taxi_fare"`
	DataSource        string `json:"data_source"`
	Status            string `json:"status"`
	Message           string `json:"message"`
}