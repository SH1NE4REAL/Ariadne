package model

type AgentStep struct {
	ToolName    string `json:"tool_name"`
	Description string `json:"description"`
}

type TripAgentResult struct {
	NeedClarification bool
	Clarification     ClarificationResult
	FinalPlan         FinalTripPlan
}