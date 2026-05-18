package agent

import (
	"context"
	"fmt"

	"ariadne/internal/llm"
	"ariadne/internal/model"
	"ariadne/internal/parser"
	"ariadne/internal/tools"
	"ariadne/internal/validator"
)

type TripAgent struct {
	TransportTool      tools.TransportTool
	AttractionTool     tools.AttractionTool
	RouteTool          tools.RouteTool
	LinkTool           tools.LinkTool
	PriceCompareTool   tools.PriceCompareTool
	BudgetTool         tools.BudgetTool
	HotelTool          tools.HotelTool
	GeoTool            tools.GeoTool
	RouteDistanceTool  tools.RouteDistanceTool
	POITool            tools.POITool
	DistanceMatrixTool tools.DistanceMatrixTool
	RouteOptimizerTool tools.RouteOptimizerTool
	FlyAIHotelTool     tools.FlyAIHotelTool
	FlyAITrainTool     tools.FlyAITrainTool
}

func NewTripAgent() TripAgent {
	return TripAgent{
		TransportTool:      tools.NewTransportTool(),
		AttractionTool:     tools.NewAttractionTool(),
		RouteTool:          tools.NewRouteTool(),
		LinkTool:           tools.NewLinkTool(),
		PriceCompareTool:   tools.NewPriceCompareTool(),
		BudgetTool:         tools.NewBudgetTool(),
		HotelTool:          tools.NewHotelTool(),
		GeoTool:            tools.NewGeoTool(),
		RouteDistanceTool:  tools.NewRouteDistanceTool(),
		POITool:            tools.NewPOITool(),
		DistanceMatrixTool: tools.NewDistanceMatrixTool(),
		RouteOptimizerTool: tools.NewRouteOptimizerTool(),
		FlyAIHotelTool:     tools.NewFlyAIHotelTool(),
		FlyAITrainTool:     tools.NewFlyAITrainTool(),
	}
}

func (a TripAgent) Run(message string, llmConfig model.LLMConfig, mapConfig model.MapConfig) model.TripAgentResult {
	_ = mapConfig
	ctx := context.Background()
	agentSteps := make([]model.AgentStep, 0)

	useLLMParser := llm.IsLLMConfigComplete(llmConfig)

	var llmClient *llm.LLMClient

	if useLLMParser {
		client, err := llm.NewLLMClient(ctx, llmConfig)
		if err != nil {
			useLLMParser = false

			agentSteps = append(agentSteps, model.AgentStep{
				ToolName:    "llm_config_checker",
				Description: "检测到 LLM 配置，但创建 Eino ChatModel 失败，已回退到规则解析器",
			})
		} else {
			llmClient = client

			agentSteps = append(agentSteps, model.AgentStep{
				ToolName:    "llm_config_checker",
				Description: "检测到完整 LLM 配置，已创建 Eino ChatModel",
			})
		}
	} else {
		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    "llm_config_checker",
			Description: "未检测到完整 LLM 配置，当前使用规则解析器进行基础解析",
		})
	}

	var tripRequest model.TripRequest

	if useLLMParser {
		llmTripRequest, err := llm.ParseTripRequestWithLLM(ctx, message, llmClient)
		if err != nil {
			tripRequest = parser.ParseTripRequest(message)

			agentSteps = append(agentSteps, model.AgentStep{
				ToolName:    "llm_parser",
				Description: "Eino 模型解析失败，已自动回退到规则解析器",
			})
		} else {
			tripRequest = llmTripRequest

			agentSteps = append(agentSteps, model.AgentStep{
				ToolName:    "llm_parser",
				Description: "使用 Eino ChatModel 和用户自带 Key 解析旅行需求",
			})
		}
	} else {
		tripRequest = parser.ParseTripRequest(message)

		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    "rule_parser",
			Description: "使用正则和关键词规则解析用户输入，提取出发地、目的地、天数、预算和偏好",
		})
	}

	questions := validator.CheckTripRequest(tripRequest)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    "request_validator",
		Description: "检查旅行请求是否完整",
	})

	if len(questions) > 0 {
		return model.TripAgentResult{
			NeedClarification: true,
			Clarification: model.ClarificationResult{
				NeedClarification: true,
				Questions:         questions,
				AgentSteps:        agentSteps,
			},
		}
	}

	var originLocation model.Location
	var destinationLocation model.Location

if mapConfig.TencentMapKey != "" {
	origin, err := a.GeoTool.Run(tripRequest.Origin, mapConfig)
	if err != nil {
		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    a.GeoTool.Name,
			Description: "出发地地理编码失败：" + err.Error(),
		})
	} else {
		originLocation = origin
		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    a.GeoTool.Name,
			Description: "成功解析出发地经纬度：" + tripRequest.Origin,
		})
	}

	
	destination, err := a.GeoTool.Run(tripRequest.Destination, mapConfig)
	if err != nil {
		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    a.GeoTool.Name,
			Description: "目的地地理编码失败：" + err.Error(),
		})
	} else {
		destinationLocation = destination
		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    a.GeoTool.Name,
			Description: "成功解析目的地经纬度：" + tripRequest.Destination,
		})
	}
} else {
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.GeoTool.Name,
		Description: "未提供腾讯位置服务 Key，跳过地理编码",
	})
}

	transportPlans := []model.TransportPlan{}
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    "transport_tool",
		Description: "旧版模拟交通方案已停用，改用 FlyAI 真实火车票/机票数据",
	})

	attractions := make([]model.Attraction, 0)

if mapConfig.TencentMapKey != "" {
	attractions = a.POITool.Run(tripRequest, mapConfig)

	if len(attractions) > 0 {
		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    a.POITool.Name,
			Description: "使用腾讯位置服务地点搜索获取真实景点 POI",
		})
	} else {
		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    a.POITool.Name,
			Description: "腾讯地点搜索未返回可用景点 POI",
		})
	}
} else {
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.POITool.Name,
		Description: "未提供腾讯位置服务 Key，跳过真实 POI 搜索",
	})
}

	dailyRoutes := a.RouteTool.Run(tripRequest, attractions)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.RouteTool.Name,
		Description: "根据旅行请求和景点列表生成每日行程路线",
	})

	bookingLinks := a.LinkTool.Run(tripRequest)
	agentSteps = append(agentSteps, model.AgentStep{
	ToolName:    a.LinkTool.Name,
	Description: "生成交通、酒店、地图和景点查询跳转链接",
	})

	bestBookingOption := model.BestBookingOption{}
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    "price_compare_tool",
		Description: "旧版模拟比价已停用，等待真实票务结果参与推荐",
	})

	budgetBreakdown := a.BudgetTool.Run(tripRequest, bestBookingOption)
	agentSteps = append(agentSteps, model.AgentStep{
	ToolName:    a.BudgetTool.Name,
	Description: "根据总预算、天数、偏好和最优交通方案拆分旅行预算",
	})

	hotelOptions := a.HotelTool.Run(tripRequest, budgetBreakdown)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.HotelTool.Name,
		Description: "根据目的地、旅行天数和每晚住宿预算推荐住宿档位",
	})

	hotelOffers := a.FlyAIHotelTool.Run(tripRequest, budgetBreakdown, attractions)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.FlyAIHotelTool.Name,
		Description: "使用 FlyAI / 飞猪真实酒店商品数据搜索酒店报价",
	})

	trainOffers := a.FlyAITrainTool.Run(tripRequest)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.FlyAITrainTool.Name,
		Description: "使用 FlyAI / 飞猪真实火车票数据搜索车次和票价",
	})

	routeDistance := a.RouteDistanceTool.Run(
	tripRequest.Origin,
	tripRequest.Destination,
	originLocation,
	destinationLocation,
	mapConfig,
)

if routeDistance.Status == "ok" {
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.RouteDistanceTool.Name,
		Description: "使用腾讯位置服务获取真实驾车距离和预计时间",
	})
} else {
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.RouteDistanceTool.Name,
		Description: "真实路线距离获取失败：" + routeDistance.Message,
	})
}

if mapConfig.TencentMapKey != "" {
	dailyRoutes = a.RouteOptimizerTool.Run(dailyRoutes, mapConfig)

	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.RouteOptimizerTool.Name,
		Description: "使用腾讯距离矩阵按最近邻策略优化每日景点顺序，并生成真实路段距离和时间",
	})
} else {
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.RouteOptimizerTool.Name,
		Description: "未提供腾讯位置服务 Key，跳过每日景点路线优化",
	})
}

	totalCost := calculateTotalCost(hotelOffers, trainOffers)
	summary := generateSummary(tripRequest, totalCost)

	finalPlan := model.FinalTripPlan{
	Request:             tripRequest,
	OriginLocation:      originLocation,
	DestinationLocation: destinationLocation,
	RouteDistance:       routeDistance,
	TransportPlans:      transportPlans,
	Attractions:         attractions,
	DailyRoutes:         dailyRoutes,
	BookingLinks:        bookingLinks,
	BestBookingOption:   bestBookingOption,
	BudgetBreakdown:     budgetBreakdown,
	HotelOptions:        hotelOptions,
	AgentSteps:          agentSteps,
	TotalEstimatedCost:  totalCost,
	Summary:             summary,
	HotelOffers:         hotelOffers,
	TrainOffers:         trainOffers,
}
	if useLLMParser && llmClient != nil {
		llmSummary, err := llm.GenerateTripSummaryWithLLM(ctx, finalPlan, llmClient)
		if err != nil {
			agentSteps = append(agentSteps, model.AgentStep{
				ToolName:    "llm_summary_generator",
				Description: "Eino 总结生成失败，已保留规则版总结",
			})
		} else {
			finalPlan.Summary = llmSummary

			agentSteps = append(agentSteps, model.AgentStep{
				ToolName:    "llm_summary_generator",
				Description: "使用 Eino ChatModel 和用户自带 Key 生成旅行总结",
			})
		}

		finalPlan.AgentSteps = agentSteps
	}

	return model.TripAgentResult{
		NeedClarification: false,
		FinalPlan:         finalPlan,
	}
}

func calculateTotalCost(hotelOffers []model.HotelOffer, trainOffers []model.TrainOffer) int {
	total := 0

	for _, offer := range hotelOffers {
		if offer.Status == "ok" && offer.TotalPrice > 0 {
			total += offer.TotalPrice
			break
		}
	}

	for _, offer := range trainOffers {
		if offer.Status == "ok" && offer.Price > 0 {
			total += offer.Price
			break
		}
	}

	return total
}

func getHotelCost(hotelOptions []model.HotelOption, hotelOffers []model.HotelOffer) int {
	for _, offer := range hotelOffers {
		if offer.Status == "ok" && offer.TotalPrice > 0 {
			return offer.TotalPrice
		}
	}

	if len(hotelOptions) > 0 {
		return hotelOptions[0].TotalPrice
	}

	return 0
}

func generateSummary(request model.TripRequest, totalCost int) string {
	if request.Budget > 0 && totalCost > request.Budget {
		return fmt.Sprintf("当前方案预估总花费约 %d 元，可能超过你的预算 %d 元，建议降低住宿档位、减少付费景点或选择更经济的交通方式。", totalCost, request.Budget)
	}

	if request.Preference == "轻松" {
		return fmt.Sprintf("当前方案预估总花费约 %d 元，已综合推荐交通、住宿和景点安排，整体节奏较轻松，适合不想太赶的旅行。", totalCost)
	}

	if request.Preference == "省钱" {
		return fmt.Sprintf("当前方案预估总花费约 %d 元，已根据省钱偏好优先控制交通和住宿成本。", totalCost)
	}

	return fmt.Sprintf("当前方案预估总花费约 %d 元，可作为初版旅行计划参考。", totalCost)
}