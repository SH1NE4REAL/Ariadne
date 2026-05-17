# Ariadne

Ariadne 是一个基于 Go + Eino 的 AI 旅游路线规划 Agent 项目。

项目目标是：用户输入自然语言旅行需求，例如出发地、目的地、旅行天数、预算和偏好，Ariadne 通过 Agent 流程解析需求、检查信息完整性、调用多个工具生成交通方案、景点推荐、每日路线，并返回结构化 JSON 结果。

当前版本支持 BYOK（Bring Your Own Key）模式，用户可以传入自己的 OpenAI-compatible 模型配置，用于大模型解析和总结生成。

---

## 当前功能

- 支持 HTTP API 调用
- 支持自然语言旅行需求输入
- 支持 BYOK 模型配置
- 支持 Eino ChatModel 调用
- 支持 LLM 解析旅行需求
- 支持信息缺失时自动追问
- 支持生成交通方案
- 支持生成景点推荐
- 支持生成每日行程
- 支持 LLM 生成自然语言总结
- 支持 Agent 执行步骤追踪

---

## 当前架构

```text
用户请求
  ↓
cmd/server/main.go
  ↓
handler.PlanTripHandler
  ↓
service.RunTripAgent
  ↓
agent.TripAgent.Run
  ↓
llm.LLMClient
  ↓
llm.ParseTripRequestWithLLM / rule parser
  ↓
validator.CheckTripRequest
  ↓
TransportTool / AttractionTool / RouteTool
  ↓
llm.GenerateTripSummaryWithLLM
  ↓
FinalTripPlan
  ↓
ApiResponse JSON