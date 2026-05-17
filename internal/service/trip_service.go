package service

import (
	"ariadne/internal/agent"
	"ariadne/internal/model"
)

func BuildFinalTripPlan(request model.TripRequest) model.FinalTripPlan {
	tripAgent := agent.NewTripAgent()
	return tripAgent.Plan(request)
}