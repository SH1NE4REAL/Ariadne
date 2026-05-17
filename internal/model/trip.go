package model

type TripRequest struct {
	RawInput    string `json:"raw_input"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Days        int    `json:"days"`
	Budget      int    `json:"budget"`
	Preference  string `json:"preference"`
}