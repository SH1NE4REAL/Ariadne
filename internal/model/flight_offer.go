package model

type FlightSegment struct {
	FlightNo      string `json:"flight_no"`
	Airline       string `json:"airline"`
	SeatClassName string `json:"seat_class_name"`

	DepCityName    string `json:"dep_city_name"`
	DepAirportName string `json:"dep_airport_name"`
	DepTerminal    string `json:"dep_terminal"`
	DepDateTime    string `json:"dep_date_time"`

	ArrCityName    string `json:"arr_city_name"`
	ArrAirportName string `json:"arr_airport_name"`
	ArrTerminal    string `json:"arr_terminal"`
	ArrDateTime    string `json:"arr_date_time"`

	DurationMinutes int `json:"duration_minutes"`
}

type FlightJourney struct {
	Direction            string          `json:"direction"` // outbound / return
	JourneyType          string          `json:"journey_type"`
	TotalDurationMinutes int             `json:"total_duration_minutes"`
	Segments             []FlightSegment `json:"segments"`
}

type FlightOffer struct {
	Provider             string          `json:"provider"`
	Price                int             `json:"price"`
	TotalDurationMinutes int             `json:"total_duration_minutes"`
	BookingLink          string          `json:"booking_link"`
	Journeys             []FlightJourney `json:"journeys"`
	DataSource           string          `json:"data_source"`
	Status               string          `json:"status"`
	Message              string          `json:"message"`
}