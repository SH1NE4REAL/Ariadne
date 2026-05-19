package model

type FinalTripPlan struct {
	Request                       TripRequest                `json:"request"`
	OriginLocation                Location                   `json:"origin_location"`
	DestinationLocation           Location                   `json:"destination_location"`
	RouteDistance                 RouteDistance              `json:"route_distance"`
	TransportPlans                []TransportPlan            `json:"transport_plans"`
	Attractions                   []Attraction               `json:"attractions"`
	TripPOIs                      []TripPOI                  `json:"trip_pois"`
	POIDebugReport                POIDebugReport             `json:"poi_debug_report"`
	DailyRoutes                   []DailyRoute               `json:"daily_routes"`
	BookingLinks                  []BookingLink              `json:"booking_links"`
	BestBookingOption             BestBookingOption          `json:"best_booking_option"`
	BudgetBreakdown               BudgetBreakdown            `json:"budget_breakdown"`
	HotelOptions                  []HotelOption              `json:"hotel_options"`
	HotelOffers                   []HotelOffer               `json:"hotel_offers"`
	TrainOffers                   []TrainOffer               `json:"train_offers"`
	PoiOffers                     []PoiOffer                 `json:"poi_offers"`
	FlightOffers                  []FlightOffer              `json:"flight_offers"`
	RecommendedTrainOffer         TrainOffer                 `json:"recommended_train_offer"`
	OutboundTrainOffers           []TrainOffer               `json:"outbound_train_offers"`
	ReturnTrainOffers             []TrainOffer               `json:"return_train_offers"`
	RecommendedOutboundTrainOffer TrainOffer                 `json:"recommended_outbound_train_offer"`
	RecommendedReturnTrainOffer   TrainOffer                 `json:"recommended_return_train_offer"`
	RecommendedFlightOffer        FlightOffer                `json:"recommended_flight_offer"`
	PreferenceConstraints         []PreferenceConstraint     `json:"preference_constraints"`
	EffectivePreferenceProfile    EffectivePreferenceProfile `json:"effective_preference_profile"`
	RecommendationViolations      []RecommendationViolation  `json:"recommendation_violations"`
	PlanQualityReport             PlanQualityReport          `json:"plan_quality_report"`
	AgentSteps                    []AgentStep                `json:"agent_steps"`
	TotalEstimatedCost            int                        `json:"total_estimated_cost"`
	TripRecommendation            TripRecommendation         `json:"trip_recommendation"`
	Summary                       string                     `json:"summary"`
}
