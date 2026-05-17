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
	AgentSteps          []AgentStep       `json:"agent_steps"`
	TotalEstimatedCost  int               `json:"total_estimated_cost"`
	Summary             string            `json:"summary"`
}