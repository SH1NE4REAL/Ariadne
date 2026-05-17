package service

import (
	"ariadne/internal/agent"
	"ariadne/internal/model"
)

func RunTripAgent(message string, llmConfig model.LLMConfig, mapConfig model.MapConfig) model.TripAgentResult {
	tripAgent := agent.NewTripAgent()
	return tripAgent.Run(message, llmConfig, mapConfig)
}