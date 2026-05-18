package tools

import (
	"context"

	"ariadne/internal/model"
	"ariadne/internal/provider/flyai"
)

type FlyAITrainTool struct {
	Name        string
	Description string
	Client      flyai.Client
}

func NewFlyAITrainTool() FlyAITrainTool {
	return FlyAITrainTool{
		Name:        "flyai_train_tool",
		Description: "使用 FlyAI / 飞猪真实火车票数据搜索车次和票价",
		Client:      flyai.NewClient(),
	}
}

func (t FlyAITrainTool) Run(request model.TripRequest) []model.TrainOffer {
	if request.StartDate == "" {
		return []model.TrainOffer{
			{
				Provider:   "fliggy",
				DataSource: "flyai_fliggy",
				Status:     "unavailable",
				Message:    "缺少出发日期，无法查询真实火车票价格。",
			},
		}
	}

	seatClassName := ""

	offers, err := t.Client.SearchTrains(
		context.Background(),
		request.Origin,
		request.Destination,
		request.StartDate,
		seatClassName,
		0,
	)

	if err != nil {
		return []model.TrainOffer{
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

func chooseTrainSeatClass(request model.TripRequest) string {
	// 第一版默认查二等座，因为普通用户最常用，也和你目前测试命令返回相符。
	if request.TransportPreference == "高铁" || request.TransportPreference == "动车" || request.TransportPreference == "火车" {
		return "second class"
	}

	return "second class"
}