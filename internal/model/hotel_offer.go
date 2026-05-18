package model

type HotelOffer struct {
	Provider      string  `json:"provider"`
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	Star          string  `json:"star"`
	PricePerNight int     `json:"price_per_night"`
	TotalPrice    int     `json:"total_price"`
	Nights        int     `json:"nights"`
	BookingLink   string  `json:"booking_link"`
	ImageURL       string  `json:"image_url"`
	NearbyPOI     string  `json:"nearby_poi"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	DataSource    string  `json:"data_source"`
	Status        string  `json:"status"`
	Message       string  `json:"message"`
}