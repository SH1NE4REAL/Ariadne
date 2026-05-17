package model

type TransportPlan struct {
	Method      string `json:"method"`
	Duration    string `json:"duration"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	BookingLink string `json:"booking_link"`
	Reason      string `json:"reason"`
}