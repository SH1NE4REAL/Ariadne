package model

type PoiOffer struct {
	Provider       string  `json:"provider"`
	Name           string  `json:"name"`
	Address        string  `json:"address"`
	Category       string  `json:"category"`
	Description    string  `json:"description"`
	FreePoiStatus  string  `json:"free_poi_status"`
	PoiLevel       string  `json:"poi_level"`
	BookingLink    string  `json:"booking_link"`
	ImageURL        string  `json:"image_url"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	HasTicketPrice bool    `json:"has_ticket_price"`
	TicketPrice    int     `json:"ticket_price"`
	DataSource     string  `json:"data_source"`
	Status         string  `json:"status"`
	Message        string  `json:"message"`
}