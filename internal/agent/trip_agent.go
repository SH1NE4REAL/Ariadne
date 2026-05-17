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
	TransportTool  tools.TransportTool
	AttractionTool tools.AttractionTool
	RouteTool      tools.RouteTool
}

func NewTripAgent() TripAgent {
	return TripAgent{
		TransportTool:  tools.NewTransportTool(),
		AttractionTool: tools.NewAttractionTool(),
		RouteTool:      tools.NewRouteTool(),
	}
}

func (a TripAgent) Run(message string, llmConfig model.LLMConfig) model.TripAgentResult {
	agentSteps := make([]model.AgentStep, 0)

	useLLMParser := isLLMConfigComplete(llmConfig)

	if useLLMParser {
		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    "llm_config_checker",
			Description: "检测到完整 LLM 配置，后续可使用用户自带 Key 进行智能解析",
		})
	} else {
		agentSteps = append(agentSteps, model.AgentStep{
			ToolName:    "llm_config_checker",
			Description: "未检测到完整 LLM 配置，当前使用规则解析器进行基础解析",
		})
	}

	

	var tripRequest model.TripRequest

if useLLMParser {
	llmTripRequest, err := llm.ParseTripRequestWithLLM(context.Background(), message, llmConfig)
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

	transportPlans := a.TransportTool.Run(tripRequest)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.TransportTool.Name,
		Description: "根据出发地、目的地、预算和偏好生成交通方案",
	})

	attractions := a.AttractionTool.Run(tripRequest)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.AttractionTool.Name,
		Description: "根据目的地、旅行天数和用户偏好推荐景点",
	})

	dailyRoutes := a.RouteTool.Run(tripRequest, attractions)
	agentSteps = append(agentSteps, model.AgentStep{
		ToolName:    a.RouteTool.Name,
		Description: "根据旅行请求和景点列表生成每日行程路线",
	})

	totalCost := calculateTotalCost(transportPlans, dailyRoutes)
	summary := generateSummary(tripRequest, totalCost)

	finalPlan := model.FinalTripPlan{
	Request:            tripRequest,
	TransportPlans:     transportPlans,
	Attractions:        attractions,
	DailyRoutes:        dailyRoutes,
	AgentSteps:         agentSteps,
	TotalEstimatedCost: totalCost,
	Summary:            summary,
}

if useLLMParser {
	llmSummary, err := llm.GenerateTripSummaryWithLLM(context.Background(), finalPlan, llmConfig)
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

func calculateTotalCost(transportPlans []model.TransportPlan, dailyRoutes []model.DailyRoute) int {
	total := 0

	if len(transportPlans) > 0 {
		total += transportPlans[0].Price
	}

	for _, route := range dailyRoutes {
		total += route.EstimatedCost
	}

	return total
}

func generateSummary(request model.TripRequest, totalCost int) string {
	if request.Budget > 0 && totalCost > request.Budget {
		return fmt.Sprintf("当前方案预估花费约 %d 元，可能超过你的预算 %d 元，建议减少高消费项目或选择更经济的交通方式。", totalCost, request.Budget)
	}

	if request.Preference == "轻松" {
		return fmt.Sprintf("当前方案预估花费约 %d 元，整体节奏较轻松，适合不想太赶的旅行。", totalCost)
	}

	if request.Preference == "省钱" {
		return fmt.Sprintf("当前方案预估花费约 %d 元，已尽量控制景点和交通成本。", totalCost)
	}

	return fmt.Sprintf("当前方案预估花费约 %d 元，可作为初版旅行计划参考。", totalCost)
}

func isLLMConfigComplete(config model.LLMConfig) bool {
	return config.APIKey != "" && config.Model != "" && config.BaseURL != ""
}