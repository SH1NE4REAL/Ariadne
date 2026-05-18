package model

type FinalTripPlan struct {
	Request             TripRequest       `json:"request"`
	OriginLocation      Location          `json:"origin_location"`
	DestinationLocation Location          `json:"destination_location"`
	RouteDistance       RouteDistance     `json:"route_distance"`
	TransportPlans      []TransportPlan   `json:"transport_plans"`
	Attractions         []Attraction      `json:"attractions"`
	DailyRoutes         []DailyRoute      `json:"daily_routes"`
	BookingLinks        []BookingLink     `json:"booking_links"`
	BestBookingOption   BestBookingOption `json:"best_booking_option"`
	BudgetBreakdown     BudgetBreakdown   `json:"budget_breakdown"`
	HotelOptions        []HotelOption     `json:"hotel_options"`
	HotelOffers 		[]HotelOffer 	  `json:"hotel_offers"`
	TrainOffers 		[]TrainOffer 	  `json:"train_offers"`
	RecommendedTrainOffer TrainOffer `json:"recommended_train_offer"`
	OutboundTrainOffers           []TrainOffer `json:"outbound_train_offers"`
	ReturnTrainOffers             []TrainOffer `json:"return_train_offers"`
	RecommendedOutboundTrainOffer TrainOffer   `json:"recommended_outbound_train_offer"`
	RecommendedReturnTrainOffer   TrainOffer   `json:"recommended_return_train_offer"`
	AgentSteps          []AgentStep       `json:"agent_steps"`
	TotalEstimatedCost  int               `json:"total_estimated_cost"`
	Summary             string            `json:"summary"`
}