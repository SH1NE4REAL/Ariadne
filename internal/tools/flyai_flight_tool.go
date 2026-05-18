package tools

import (
	"context"

	"ariadne/internal/model"
	"ariadne/internal/provider/flyai"
)

type FlyAIFlightTool struct {
	Name        string
	Description string
	Client      flyai.Client
}

func NewFlyAIFlightTool() FlyAIFlightTool {
	return FlyAIFlightTool{
		Name:        "flyai_flight_tool",
		Description: "使用 FlyAI / 飞猪真实机票数据搜索航班和票价",
		Client:      flyai.NewClient(),
	}
}

func (t FlyAIFlightTool) Run(request model.TripRequest) []model.FlightOffer {
	if request.StartDate == "" {
		return []model.FlightOffer{
			{
				Provider:   "fliggy",
				DataSource: "flyai_fliggy",
				Status:     "unavailable",
				Message:    "缺少出发日期，无法查询真实机票价格。",
			},
		}
	}

	seatClassName := "economy"

	offers, err := t.Client.SearchFlights(
		context.Background(),
		request.Origin,
		request.Destination,
		request.StartDate,
		request.EndDate,
		seatClassName,
		0,
	)

	if err != nil {
		return []model.FlightOffer{
			{
				Provider:   "fliggy",
				DataSource: "flyai_fliggy",
				Status:     "unavailable",
				Message:    err.Error(),
			},
		}
	}

	return offers
}