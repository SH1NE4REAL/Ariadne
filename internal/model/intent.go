package model

type TripIntentProfile struct {
	TransportIntent  DomainIntent `json:"transport_intent"`
	HotelIntent      DomainIntent `json:"hotel_intent"`
	AttractionIntent DomainIntent `json:"attraction_intent"`
	FoodIntent       DomainIntent `json:"food_intent"`
	RouteIntent      DomainIntent `json:"route_intent"`
	ConstraintIntent DomainIntent `json:"constraint_intent"`
}

type DomainIntent struct {
	HardPreferTags []string `json:"hard_prefer_tags"`
	SoftPreferTags []string `json:"soft_prefer_tags"`
	HardAvoidTags  []string `json:"hard_avoid_tags"`
	SoftAvoidTags  []string `json:"soft_avoid_tags"`

	HardPreferKeywords []string `json:"hard_prefer_keywords"`
	SoftPreferKeywords []string `json:"soft_prefer_keywords"`
	HardAvoidKeywords  []string `json:"hard_avoid_keywords"`
	SoftAvoidKeywords  []string `json:"soft_avoid_keywords"`

	Source string `json:"source"` // current_request / memory / system_default
}
