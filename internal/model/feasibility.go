package model

type FeasibilityIssue struct {
	Level   string `json:"level"` // warning / severe / impossible
	Code    string `json:"code"`
	Message string `json:"message"`
}
