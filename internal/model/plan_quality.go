package model

type PlanQualityReport struct {
	HardConstraintPassed     bool               `json:"hard_constraint_passed"`
	MainAttractionCount      int                `json:"main_attraction_count"`
	FoodSpotCount            int                `json:"food_spot_count"`
	ShoppingSpotCount        int                `json:"shopping_spot_count"`
	InvalidPOICount          int                `json:"invalid_poi_count"`
	AvgTransferMinutes       int                `json:"avg_transfer_minutes"`
	HotelDistanceToCoreKM    float64            `json:"hotel_distance_to_core_km"`
	CoreIntentCoverage       map[string]bool    `json:"core_intent_coverage"`
	SummaryConsistencyPassed bool               `json:"summary_consistency_passed"`
	BudgetFeasibility        string             `json:"budget_feasibility"` // ok / tight / impossible / unknown
	FeasibilityIssues        []FeasibilityIssue `json:"feasibility_issues"`
	Warnings                 []string           `json:"warnings"`
	Score                    int                `json:"score"`
}
