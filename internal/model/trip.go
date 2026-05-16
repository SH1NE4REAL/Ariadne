package model

type TripRequest struct {
	RawInput    string
	Origin      string
	Destination string
	Days        int
	Budget      int
	Preference  string
}