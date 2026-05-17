package service

import (
	"ariadne/internal/agent"
	"ariadne/internal/model"
)

func RunTripAgent(message string, llmConfig model.LLMConfig) model.TripAgentResult {
	tripAgent := agent.NewTripAgent()
	return tripAgent.Run(message, llmConfig)
}