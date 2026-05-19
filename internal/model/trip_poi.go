package model

type TripPOI struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Address     string  `json:"address"`
	Description string  `json:"description"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`

	Role string   `json:"role"` // main_attraction / food_spot / shopping_spot / transit_nearby / invalid
	Tags []string `json:"tags"`

	Score int `json:"score"`

	Attraction Attraction `json:"attraction"`
}

type POICluster struct {
	CenterLat float64   `json:"center_lat"`
	CenterLng float64   `json:"center_lng"`
	POIs      []TripPOI `json:"pois"`
	Tags      []string  `json:"tags"`
	Score     int       `json:"score"`
}

type DayPlanTemplate struct {
	DayType            string `json:"day_type"` // arrival / full_day / departure
	MainPOICount       int    `json:"main_poi_count"`
	FoodPOICount       int    `json:"food_poi_count"`
	ShoppingPOICount   int    `json:"shopping_poi_count"`
	AllowNightView     bool   `json:"allow_night_view"`
	MaxTransferMinutes int    `json:"max_transfer_minutes"`
}
