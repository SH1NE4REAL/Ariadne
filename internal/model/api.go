package model

type TripPlanRequest struct {
	Message   string    `json:"message"`
	LLMConfig LLMConfig `json:"llm_config"`
	MapConfig MapConfig `json:"map_config"`
}

type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}