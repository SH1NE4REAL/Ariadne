package model

type RecommendationCostItem struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Amount     int    `json:"amount"`
	Currency   string `json:"currency"`
	DataSource string `json:"data_source"`
}

type TripRecommendation struct {
	RecommendedTransportType string      `json:"recommended_transport_type"` // flight / train / unknown
	RecommendedHotel         HotelOffer  `json:"recommended_hotel"`
	RecommendedOutboundTrain TrainOffer  `json:"recommended_outbound_train"`
	RecommendedReturnTrain   TrainOffer  `json:"recommended_return_train"`
	RecommendedFlight        FlightOffer `json:"recommended_flight"`

	TotalRealCost int                      `json:"total_real_cost"`
	Budget        int                      `json:"budget"`
	BudgetStatus  string                   `json:"budget_status"` // ok / over_budget / unknown
	OverBudget    int                      `json:"over_budget"`
	CostItems     []RecommendationCostItem `json:"cost_items"`

	Reason string `json:"reason"`
}