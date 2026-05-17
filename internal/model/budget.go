package model

type BudgetBreakdown struct {
	TotalBudget            int      `json:"total_budget"`
	TransportBudget        int      `json:"transport_budget"`
	HotelBudget            int      `json:"hotel_budget"`
	HotelBudgetPerNight    int      `json:"hotel_budget_per_night"`
	FoodBudget             int      `json:"food_budget"`
	AttractionBudget       int      `json:"attraction_budget"`
	ReserveBudget          int      `json:"reserve_budget"`
	EstimatedTransportCost int      `json:"estimated_transport_cost"`
	RemainingBudget        int      `json:"remaining_budget"`
	Status                 string   `json:"status"`
	Suggestions            []string `json:"suggestions"`
}