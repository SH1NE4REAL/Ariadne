package model

type TripRequest struct {
	RawInput             string `json:"raw_input"`
	Origin               string `json:"origin"`
	Destination          string `json:"destination"`
	Days                 int    `json:"days"`
	Budget               int    `json:"budget"`
	Preference           string `json:"preference"`
	TransportPreference  string `json:"transport_preference"`
	LocalTransportMode    string `json:"local_transport_mode"`
	StartDate            string `json:"start_date"`
	EndDate              string `json:"end_date"`
}