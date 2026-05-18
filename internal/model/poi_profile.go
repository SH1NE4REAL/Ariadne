package model

type POIProfile struct {
	Name        string
	Category    string
	Description string
	Address     string

	Tags          []string
	Invalid       bool
	InvalidReason string
	Score         int

	Attraction Attraction
}
