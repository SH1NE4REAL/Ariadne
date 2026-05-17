package service

import (
	"ariadne/internal/agent"
	"ariadne/internal/model"
)

func RunTripAgent(message string) model.TripAgentResult {
	tripAgent := agent.NewTripAgent()
	return tripAgent.Run(message)
}