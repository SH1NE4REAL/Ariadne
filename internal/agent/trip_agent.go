package agent

import (
	"context"
	"fmt"
	"strings"
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
	FlyAIPoiTool 		tools.FlyAIPoiTool
	FlyAIFlightTool tools.FlyAIFlightTool
	RouteFeasibilityTool tools.RouteFeasibilityTool
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
		FlyAIPoiTool: 		tools.NewFlyAIPoiTool(),
		FlyAIFlightTool: tools.NewFlyAIFlightTool(),
		RouteFeasibilityTool: tools.NewRouteFeasibilityTool(),
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

	outboundTrainOffers := a.FlyAITrainTool.RunOutbound(tripRequest)
	recommendedOutboundTrainOffer := selectRecommendedTrainOffer(tripRequest, outboundTrainOffers, "outbound")

	returnTrainOffers := a.FlyAITrainTool.RunReturn(tripRequest)
	recommendedReturnTrainOffer := selectRecommendedTrainOffer(tripRequest, returnTrainOffers, "return")

	// 兼容旧字段：train_offers 先继续等于去程结果
	trainOffers := outboundTrainOffers
	recommendedTrainOffer := recommendedOutboundTrainOffer

	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.FlyAITrainTool.Name,
		Description: "使用 FlyAI / 飞猪真实火车票数据搜索去程和返程车次与票价",
	})
	

	flightOffers := a.FlyAIFlightTool.Run(tripRequest)

	recommendedFlightOffer := selectRecommendedFlightOffer(tripRequest, flightOffers)


	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.FlyAIFlightTool.Name,
		Description: "使用 FlyAI / 飞猪真实机票数据搜索航班和票价",
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
	
	dailyRoutes = a.RouteFeasibilityTool.Run(tripRequest, dailyRoutes)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.RouteFeasibilityTool.Name,
		Description: "根据真实路段距离、通勤时间和用户偏好过滤不可行路线",
})
} else {
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.RouteOptimizerTool.Name,
		Description: "未提供腾讯位置服务 Key，跳过每日景点路线优化",
	})
}

	poiOffers := a.FlyAIPoiTool.Run(tripRequest, attractions)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.FlyAIPoiTool.Name,
		Description: "使用 FlyAI / 飞猪补充景点详情、图片和跳转链接",
	})

	tripRecommendation := buildTripRecommendation(
		tripRequest,
		budgetBreakdown,
		hotelOffers,
		recommendedOutboundTrainOffer,
		recommendedReturnTrainOffer,
		recommendedFlightOffer,
	)

	totalCost := tripRecommendation.TotalRealCost
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
	RecommendedTrainOffer:         recommendedTrainOffer,
	OutboundTrainOffers:           outboundTrainOffers,
	ReturnTrainOffers:             returnTrainOffers,
	RecommendedOutboundTrainOffer: recommendedOutboundTrainOffer,
	RecommendedReturnTrainOffer:   recommendedReturnTrainOffer,
	PoiOffers: poiOffers,
	FlightOffers:           flightOffers,
	RecommendedFlightOffer: recommendedFlightOffer,
	TripRecommendation: tripRecommendation,
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

func calculateTotalCost(
	request model.TripRequest,
	hotelOffers []model.HotelOffer,
	recommendedOutboundTrainOffer model.TrainOffer,
	recommendedReturnTrainOffer model.TrainOffer,
	recommendedFlightOffer model.FlightOffer,
) int {
	total := 0

	for _, offer := range hotelOffers {
		if offer.Status == "ok" && offer.TotalPrice > 0 {
			total += offer.TotalPrice
			break
		}
	}

	if request.TransportPreference == "飞机" {
		if recommendedFlightOffer.Status == "ok" && recommendedFlightOffer.Price > 0 {
			total += recommendedFlightOffer.Price
		}
		return total
	}

	if recommendedOutboundTrainOffer.Status == "ok" && recommendedOutboundTrainOffer.Price > 0 {
		total += recommendedOutboundTrainOffer.Price
	}

	if recommendedReturnTrainOffer.Status == "ok" && recommendedReturnTrainOffer.Price > 0 {
		total += recommendedReturnTrainOffer.Price
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
func selectRecommendedTrainOffer(request model.TripRequest, trainOffers []model.TrainOffer, direction string) model.TrainOffer {
	candidates := make([]model.TrainOffer, 0)

	for _, offer := range trainOffers {
		if offer.Status != "ok" || offer.Price <= 0 {
			continue
		}

		if request.TransportPreference == "高铁" {
			if !isHighSpeedTrainOffer(offer) {
				continue
			}
		}

		candidates = append(candidates, offer)
	}

	if len(candidates) == 0 {
		return model.TrainOffer{
			Provider:   "fliggy",
			Direction:  direction,
			DataSource: "flyai_fliggy",
			Status:     "unavailable",
			Message:    "没有找到符合用户交通偏好的真实火车票结果。",
		}
	}

	best := candidates[0]

	for _, offer := range candidates[1:] {
		if isBetterTrainOffer(request, offer, best) {
			best = offer
		}
	}

	return best
}

func isHighSpeedTrainOffer(offer model.TrainOffer) bool {
	for _, segment := range offer.Segments {
		if segment.TrainType == "高铁" || segment.TrainType == "动车" || segment.TrainType == "城际" {
			return true
		}

		if len(segment.TrainNo) > 0 {
			first := segment.TrainNo[0]
			if first == 'G' || first == 'D' || first == 'C' {
				return true
			}
		}
	}

	return false
}

func isBetterTrainOffer(request model.TripRequest, current model.TrainOffer, best model.TrainOffer) bool {
	// 轻松偏好：优先直达，再看耗时，再看价格
	if request.Preference == "轻松" {
		if current.JourneyType == "直达" && best.JourneyType != "直达" {
			return true
		}

		if current.JourneyType != "直达" && best.JourneyType == "直达" {
			return false
		}

		if current.TotalDurationMinutes > 0 && best.TotalDurationMinutes > 0 {
			if current.TotalDurationMinutes < best.TotalDurationMinutes {
				return true
			}
		}

		return current.Price < best.Price
	}

	// 省钱偏好：优先低价
	if request.Preference == "省钱" {
		return current.Price < best.Price
	}

	// 默认：价格更低优先
	return current.Price < best.Price
}

func selectRecommendedFlightOffer(request model.TripRequest, flightOffers []model.FlightOffer) model.FlightOffer {
	candidates := make([]model.FlightOffer, 0)

	for _, offer := range flightOffers {
		if offer.Status != "ok" || offer.Price <= 0 {
			continue
		}

		candidates = append(candidates, offer)
	}

	if len(candidates) == 0 {
		return model.FlightOffer{
			Provider:   "fliggy",
			DataSource: "flyai_fliggy",
			Status:     "unavailable",
			Message:    "没有找到可用的真实机票结果。",
		}
	}

	best := candidates[0]

	for _, offer := range candidates[1:] {
		if isBetterFlightOffer(request, offer, best) {
			best = offer
		}
	}

	return best
}

func isBetterFlightOffer(request model.TripRequest, current model.FlightOffer, best model.FlightOffer) bool {
	// 轻松：优先总时长短，再看价格
	if request.Preference == "轻松" {
		if current.TotalDurationMinutes > 0 && best.TotalDurationMinutes > 0 {
			if current.TotalDurationMinutes < best.TotalDurationMinutes {
				return true
			}
		}

		return current.Price < best.Price
	}

	// 省钱：优先低价
	if request.Preference == "省钱" {
		return current.Price < best.Price
	}

	// 默认：优先低价
	return current.Price < best.Price
}

func buildTripRecommendation(
	request model.TripRequest,
	budgetBreakdown model.BudgetBreakdown,
	hotelOffers []model.HotelOffer,
	recommendedOutboundTrain model.TrainOffer,
	recommendedReturnTrain model.TrainOffer,
	recommendedFlight model.FlightOffer,
) model.TripRecommendation {
	recommendedHotel := selectRecommendedHotel(request, budgetBreakdown, hotelOffers)
	transportType := selectRecommendedTransportType(request, recommendedOutboundTrain, recommendedReturnTrain, recommendedFlight)

	costItems := make([]model.RecommendationCostItem, 0)
	totalRealCost := 0

	if recommendedHotel.Status == "ok" && recommendedHotel.TotalPrice > 0 {
		costItems = append(costItems, model.RecommendationCostItem{
			Type:       "hotel",
			Name:       recommendedHotel.Name,
			Amount:     recommendedHotel.TotalPrice,
			Currency:   "CNY",
			DataSource: recommendedHotel.DataSource,
		})
		totalRealCost += recommendedHotel.TotalPrice
	}

	if transportType == "flight" {
		if recommendedFlight.Status == "ok" && recommendedFlight.Price > 0 {
			costItems = append(costItems, model.RecommendationCostItem{
				Type:       "flight_round_trip",
				Name:       "往返机票",
				Amount:     recommendedFlight.Price,
				Currency:   "CNY",
				DataSource: recommendedFlight.DataSource,
			})
			totalRealCost += recommendedFlight.Price
		}
	}

	if transportType == "train" {
		if recommendedOutboundTrain.Status == "ok" && recommendedOutboundTrain.Price > 0 {
			costItems = append(costItems, model.RecommendationCostItem{
				Type:       "train_outbound",
				Name:       "去程火车票",
				Amount:     recommendedOutboundTrain.Price,
				Currency:   "CNY",
				DataSource: recommendedOutboundTrain.DataSource,
			})
			totalRealCost += recommendedOutboundTrain.Price
		}

		if recommendedReturnTrain.Status == "ok" && recommendedReturnTrain.Price > 0 {
			costItems = append(costItems, model.RecommendationCostItem{
				Type:       "train_return",
				Name:       "返程火车票",
				Amount:     recommendedReturnTrain.Price,
				Currency:   "CNY",
				DataSource: recommendedReturnTrain.DataSource,
			})
			totalRealCost += recommendedReturnTrain.Price
		}
	}

	budgetStatus := "unknown"
	overBudget := 0

	if request.Budget > 0 {
		if totalRealCost <= request.Budget {
			budgetStatus = "ok"
		} else {
			budgetStatus = "over_budget"
			overBudget = totalRealCost - request.Budget
		}
	}

	return model.TripRecommendation{
		RecommendedTransportType: transportType,
		RecommendedHotel:         recommendedHotel,
		RecommendedOutboundTrain: recommendedOutboundTrain,
		RecommendedReturnTrain:   recommendedReturnTrain,
		RecommendedFlight:        recommendedFlight,
		TotalRealCost:            totalRealCost,
		Budget:                   request.Budget,
		BudgetStatus:             budgetStatus,
		OverBudget:               overBudget,
		CostItems:                costItems,
		Reason:                   buildTripRecommendationReason(request, transportType, totalRealCost, budgetStatus, overBudget),
	}
}

func selectRecommendedTransportType(
	request model.TripRequest,
	recommendedOutboundTrain model.TrainOffer,
	recommendedReturnTrain model.TrainOffer,
	recommendedFlight model.FlightOffer,
) string {
	if request.TransportPreference == "飞机" || strings.Contains(request.RawInput, "飞机") {
		if recommendedFlight.Status == "ok" {
			return "flight"
		}
	}

	if request.TransportPreference == "高铁" ||
		request.TransportPreference == "动车" ||
		request.TransportPreference == "火车" ||
		strings.Contains(request.RawInput, "高铁") ||
		strings.Contains(request.RawInput, "火车") {
		if recommendedOutboundTrain.Status == "ok" || recommendedReturnTrain.Status == "ok" {
			return "train"
		}
	}

	if strings.Contains(request.RawInput, "快") && recommendedFlight.Status == "ok" {
		return "flight"
	}

	trainCost := 0
	if recommendedOutboundTrain.Status == "ok" {
		trainCost += recommendedOutboundTrain.Price
	}
	if recommendedReturnTrain.Status == "ok" {
		trainCost += recommendedReturnTrain.Price
	}

	if recommendedFlight.Status == "ok" && trainCost > 0 {
		if recommendedFlight.Price <= trainCost {
			return "flight"
		}
		return "train"
	}

	if recommendedFlight.Status == "ok" {
		return "flight"
	}

	if recommendedOutboundTrain.Status == "ok" || recommendedReturnTrain.Status == "ok" {
		return "train"
	}

	return "unknown"
}

func selectRecommendedHotel(
	request model.TripRequest,
	budgetBreakdown model.BudgetBreakdown,
	hotelOffers []model.HotelOffer,
) model.HotelOffer {
	candidates := make([]model.HotelOffer, 0)

	for _, offer := range hotelOffers {
		if offer.Status == "ok" && offer.TotalPrice > 0 {
			candidates = append(candidates, offer)
		}
	}

	if len(candidates) == 0 {
		return model.HotelOffer{
			Provider:   "fliggy",
			DataSource: "flyai_fliggy",
			Status:     "unavailable",
			Message:    "没有找到可用的真实酒店报价。",
		}
	}

	best := candidates[0]

	for _, offer := range candidates[1:] {
		if isBetterHotelOffer(request, budgetBreakdown, offer, best) {
			best = offer
		}
	}

	return best
}

func isBetterHotelOffer(
	request model.TripRequest,
	budgetBreakdown model.BudgetBreakdown,
	current model.HotelOffer,
	best model.HotelOffer,
) bool {
	currentWithinBudget := budgetBreakdown.HotelBudget <= 0 || current.TotalPrice <= budgetBreakdown.HotelBudget
	bestWithinBudget := budgetBreakdown.HotelBudget <= 0 || best.TotalPrice <= budgetBreakdown.HotelBudget

	if currentWithinBudget && !bestWithinBudget {
		return true
	}

	if !currentWithinBudget && bestWithinBudget {
		return false
	}

	currentNearTransit := isNearTransit(current.NearbyPOI)
	bestNearTransit := isNearTransit(best.NearbyPOI)

	if request.Preference == "轻松" || strings.Contains(request.RawInput, "方便") || strings.Contains(request.RawInput, "地铁") {
		if currentNearTransit && !bestNearTransit {
			return true
		}

		if !currentNearTransit && bestNearTransit {
			return false
		}
	}

	return current.TotalPrice < best.TotalPrice
}

func isNearTransit(nearbyPOI string) bool {
	return strings.Contains(nearbyPOI, "地铁") ||
		strings.Contains(nearbyPOI, "站") ||
		strings.Contains(nearbyPOI, "机场") ||
		strings.Contains(nearbyPOI, "火车站")
}

func buildTripRecommendationReason(
	request model.TripRequest,
	transportType string,
	totalRealCost int,
	budgetStatus string,
	overBudget int,
) string {
	transportText := "真实交通方案"
	if transportType == "flight" {
		transportText = "真实往返机票方案"
	}
	if transportType == "train" {
		transportText = "真实往返火车票方案"
	}

	if request.Budget <= 0 {
		return fmt.Sprintf("已基于 FlyAI / 飞猪真实酒店报价和%s生成推荐方案，但用户没有提供明确预算。", transportText)
	}

	if budgetStatus == "ok" {
		return fmt.Sprintf("已基于 FlyAI / 飞猪真实酒店报价和%s生成推荐方案，当前已知真实费用约 %d 元，在预算范围内。", transportText, totalRealCost)
	}

	if budgetStatus == "over_budget" {
		return fmt.Sprintf("已基于 FlyAI / 飞猪真实酒店报价和%s生成推荐方案，当前已知真实费用约 %d 元，超出预算 %d 元。", transportText, totalRealCost, overBudget)
	}

	return "推荐方案已生成，但部分真实费用数据不可用。"
}