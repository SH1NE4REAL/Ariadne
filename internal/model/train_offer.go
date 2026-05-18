package model

type TrainSegment struct {
	TrainNo       string `json:"train_no"`
	TrainType     string `json:"train_type"`
	SeatClassName string `json:"seat_class_name"`

	DepCityName    string `json:"dep_city_name"`
	DepStationName string `json:"dep_station_name"`
	DepDateTime    string `json:"dep_date_time"`

	ArrCityName    string `json:"arr_city_name"`
	ArrStationName string `json:"arr_station_name"`
	ArrDateTime    string `json:"arr_date_time"`

	DurationMinutes int `json:"duration_minutes"`
}

type TrainOffer struct {
	Provider             string         `json:"provider"`
	JourneyType          string         `json:"journey_type"`
	Price                int            `json:"price"`
	TotalDurationMinutes int            `json:"total_duration_minutes"`
	BookingLink          string         `json:"booking_link"`
	Segments             []TrainSegment `json:"segments"`
	DataSource           string         `json:"data_source"`
	Status               string         `json:"status"`
	Message              string         `json:"message"`
}