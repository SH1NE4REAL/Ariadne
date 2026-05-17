package model

type HotelOption struct {
	Name          string `json:"name"`
	Level         string `json:"level"`
	Location      string `json:"location"`
	Description   string `json:"description"`
	PricePerNight int    `json:"price_per_night"`
	Nights        int    `json:"nights"`
	TotalPrice    int    `json:"total_price"`
	BookingLink   string `json:"booking_link"`
	Reason        string `json:"reason"`
}