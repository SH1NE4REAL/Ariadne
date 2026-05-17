package model

type TripPlanRequest struct {
	Message string `json:"message"`
}

type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}