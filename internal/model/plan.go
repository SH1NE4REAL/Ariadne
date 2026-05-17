package model

type FinalTripPlan struct {
	Request            TripRequest     `json:"request"`
	TransportPlans     []TransportPlan `json:"transport_plans"`
	Attractions        []Attraction    `json:"attractions"`
	DailyRoutes        []DailyRoute    `json:"daily_routes"`
	AgentSteps         []AgentStep     `json:"agent_steps"`
	TotalEstimatedCost int             `json:"total_estimated_cost"`
	Summary            string          `json:"summary"`
}