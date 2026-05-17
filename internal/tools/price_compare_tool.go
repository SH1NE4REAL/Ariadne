package tools

import (
	"net/url"
	"sort"

	"ariadne/internal/model"
)

type PriceCompareTool struct {
	Name        string
	Description string
}

func NewPriceCompareTool() PriceCompareTool {
	return PriceCompareTool{
		Name:        "price_compare_tool",
		Description: "根据预算、偏好和候选价格，选择综合最优的购买或查询链接",
	}
}

func (t PriceCompareTool) Run(request model.TripRequest) model.BestBookingOption {
	quotes := generateMockPriceQuotes(request)
	quotes = filterQuotesByTransportPreference(request, quotes)

	for i := range quotes {
		quotes[i].Score = calculateQuoteScore(request, quotes[i])
		quotes[i].Reason = generateQuoteReason(request, quotes[i])
	}

	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].Score > quotes[j].Score
	})

	if len(quotes) == 0 {
		return model.BestBookingOption{}
	}

	best := quotes[0]

	alternatives := []model.PriceQuote{}
	if len(quotes) > 1 {
		alternatives = quotes[1:]
	}

	return model.BestBookingOption{
		Best:         best,
		Alternatives: alternatives,
	}
}

func generateMockPriceQuotes(request model.TripRequest) []model.PriceQuote {
	origin := url.QueryEscape(request.Origin)
	destination := url.QueryEscape(request.Destination)

	return []model.PriceQuote{
		{
			Type:     "transport",
			Platform: "12306",
			Method:   "高铁",
			Price:    200,
			Duration: "约1.5-3小时",
			URL:      "https://www.12306.cn/index/",
		},
		{
			Type:     "transport",
			Platform: "携程",
			Method:   "飞机",
			Price:    600,
			Duration: "约2-4小时，含机场通勤和安检时间",
			URL:      "https://flights.ctrip.com/",
		},
		{
			Type:     "transport",
			Platform: "高德地图",
			Method:   "自驾",
			Price:    400,
			Duration: "根据路况变化较大",
			URL:      "https://ditu.amap.com/dir?from=" + origin + "&to=" + destination,
		},
	}
}

func calculateQuoteScore(request model.TripRequest, quote model.PriceQuote) int {
	score := 100

	// 价格越低越加分
	if quote.Price <= 200 {
		score += 30
	} else if quote.Price <= 500 {
		score += 15
	} else {
		score -= 10
	}

	// 如果用户预算较低，更看重低价
	if request.Budget > 0 {
		transportBudget := request.Budget / 3

		if quote.Price <= transportBudget {
			score += 25
		} else {
			score -= 20
		}
	}

	// 根据偏好调整
	switch request.Preference {
	case "省钱":
		if quote.Price <= 300 {
			score += 35
		}
	case "轻松":
		if quote.Method == "高铁" {
			score += 30
		}
		if quote.Method == "飞机" {
			score -= 10
		}
	case "美食", "拍照":
		if quote.Method == "高铁" {
			score += 15
		}
	}

	// 方法本身的基础倾向
	if quote.Method == "高铁" {
		score += 20
	}

	return score
}

func generateQuoteReason(request model.TripRequest, quote model.PriceQuote) string {
	if request.TransportPreference != "" && quote.Method == request.TransportPreference {
		return "该方案符合你指定的交通方式，并在当前候选中综合价格、时间和便利性后作为优先推荐。"
	}
	
	if request.Preference == "省钱" && quote.Price <= 300 {
		return "该方案价格较低，符合省钱偏好，适合预算有限的用户。"
	}

	if request.Preference == "轻松" && quote.Method == "高铁" {
		return "高铁价格适中、时间稳定，省去了机场通勤和安检等待，更适合轻松出行。"
	}

	if request.Budget > 0 && quote.Price <= request.Budget/3 {
		return "该方案交通成本占总预算比例较低，适合当前预算。"
	}

	if quote.Method == "飞机" {
		return "飞机适合远距离出行，但价格和机场通勤成本较高，本次不一定是最优选择。"
	}

	if quote.Method == "自驾" {
		return "自驾自由度较高，但费用和时间受人数、油费、停车和路况影响较大。"
	}

	return "该方案综合价格、时间和便利性后可作为备选。"
}

func filterQuotesByTransportPreference(request model.TripRequest, quotes []model.PriceQuote) []model.PriceQuote {
	if request.TransportPreference == "" {
		return quotes
	}

	filtered := make([]model.PriceQuote, 0)

	for _, quote := range quotes {
		if quote.Method == request.TransportPreference {
			filtered = append(filtered, quote)
		}
	}

	if len(filtered) == 0 {
		return quotes
	}

	return filtered
}