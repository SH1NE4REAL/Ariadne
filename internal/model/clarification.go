package model

type ClarificationQuestion struct {
	Field    string `json:"field"`
	Question string `json:"question"`
}

type ClarificationResult struct {
	NeedClarification bool                    `json:"need_clarification"`
	Questions         []ClarificationQuestion `json:"questions"`
	AgentSteps        []AgentStep             `json:"agent_steps"`
}