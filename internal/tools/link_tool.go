package tools

import (
	"net/url"

	"ariadne/internal/model"
)

type LinkTool struct {
	Name        string
	Description string
}

func NewLinkTool() LinkTool {
	return LinkTool{
		Name:        "link_tool",
		Description: "根据旅行请求生成交通、酒店、地图和景点查询跳转链接",
	}
}

func (t LinkTool) Run(request model.TripRequest) []model.BookingLink {
	return GenerateBookingLinks(request)
}

func GenerateBookingLinks(request model.TripRequest) []model.BookingLink {
	destination := url.QueryEscape(request.Destination)
	origin := url.QueryEscape(request.Origin)

	links := []model.BookingLink{
		{
			Type:        "transport",
			Name:        "12306 火车票查询",
			Description: "用于查询高铁、动车、火车票信息。",
			URL:         "https://www.12306.cn/index/",
		},
		{
			Type:        "transport",
			Name:        "携程机票查询",
			Description: "用于查询机票价格和航班信息。",
			URL:         "https://flights.ctrip.com/",
		},
		{
			Type:        "hotel",
			Name:        "携程酒店查询",
			Description: "用于查询目的地酒店和住宿价格。",
			URL:         "https://hotels.ctrip.com/hotels/list?cityName=" + destination,
		},
		{
			Type:        "map",
			Name:        "高德地图路线查询",
			Description: "用于查看出发地到目的地的地图路线。",
			URL:         "https://ditu.amap.com/dir?from=" + origin + "&to=" + destination,
		},
		{
			Type:        "attraction",
			Name:        "高德地图景点搜索",
			Description: "用于搜索目的地附近景点、餐饮和商圈。",
			URL:         "https://ditu.amap.com/search?query=" + destination + "%20景点",
		},
	}
	return links
}