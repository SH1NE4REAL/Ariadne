package model

type BookingLink struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}