package service

import "ariadne/internal/model"

func CheckTripRequest(request model.TripRequest) []model.ClarificationQuestion {
	questions := make([]model.ClarificationQuestion, 0)

	if request.Origin == "" {
		questions = append(questions, model.ClarificationQuestion{
			Field:    "origin",
			Question: "请告诉我你的出发地是哪里？",
		})
	}

	if request.Destination == "" {
		questions = append(questions, model.ClarificationQuestion{
			Field:    "destination",
			Question: "请告诉我你的目的地是哪里？",
		})
	}

	if request.Days <= 0 {
		questions = append(questions, model.ClarificationQuestion{
			Field:    "days",
			Question: "请告诉我你打算玩几天？",
		})
	}

	if request.Budget <= 0 {
		questions = append(questions, model.ClarificationQuestion{
			Field:    "budget",
			Question: "请告诉我你的预算大概是多少？",
		})
	}

	return questions
}