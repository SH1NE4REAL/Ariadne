package model

type POIFallbackSearchGroup struct {
	Intent   string   `json:"intent"`
	Keywords []string `json:"keywords"`
}

type POIDebugReport struct {
	RawPOICount               int               `json:"raw_poi_count"`
	RawPOICountByKeyword      map[string]int    `json:"raw_poi_count_by_keyword"`
	AfterRoleClassifyCount    int               `json:"after_role_classify_count"`
	AfterHardAvoidFilterCount int               `json:"after_hard_avoid_filter_count"`
	AfterDistanceFilterCount  int               `json:"after_distance_filter_count"`
	FinalRoutablePOICount     int               `json:"final_routable_poi_count"`
	RejectedReasons           map[string]int    `json:"rejected_reasons"`
	FallbackKeywords          []string          `json:"fallback_keywords"`
	SearchStatusByKeyword     map[string]string `json:"search_status_by_keyword"`
	SearchErrorByKeyword      map[string]string `json:"search_error_by_keyword"`
}
