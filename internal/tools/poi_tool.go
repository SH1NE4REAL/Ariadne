package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ariadne/internal/model"
)

type POITool struct {
	Name        string
	Description string
}

func NewPOITool() POITool {
	return POITool{
		Name:        "poi_tool",
		Description: "使用腾讯位置服务地点搜索 API 获取真实景点 POI",
	}
}

func (t POITool) Run(request model.TripRequest, mapConfig model.MapConfig) []model.Attraction {
	keywords := buildPOIKeywords(request)

	attractions := make([]model.Attraction, 0)
	seen := make(map[string]bool)

	for _, keyword := range keywords {
		results, err := SearchPOIsWithTencent(request.Destination, keyword, mapConfig)
		if err != nil {
			continue
		}

		for _, attraction := range results {
			if seen[attraction.Name] {
				continue
			}

			seen[attraction.Name] = true
			attractions = append(attractions, attraction)

			if len(attractions) >= 8 {
				return attractions
			}
		}
	}

	return attractions
}

type tencentPlaceSearchResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Count   int    `json:"count"`
	Data    []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Address  string `json:"address"`
		Category string `json:"category"`
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"data"`
}

func SearchPOIsWithTencent(destination string, keyword string, mapConfig model.MapConfig) ([]model.Attraction, error){
	if mapConfig.TencentMapKey == "" {
		return nil, errors.New("tencent map key is empty")
	}

	if destination == "" {
		return nil, errors.New("destination is empty")
	}

	endpoint := "https://apis.map.qq.com/ws/place/v1/search"

	query := url.Values{}
	query.Set("boundary", fmt.Sprintf("region(%s,0)", destination))
	query.Set("keyword", keyword)
	query.Set("page_size", "8")
	query.Set("page_index", "1")
	query.Set("output", "json")
	query.Set("key", mapConfig.TencentMapKey)

	requestURL := endpoint + "?" + query.Encode()

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tencent poi http status: %d", resp.StatusCode)
	}

	var result tencentPlaceSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != 0 {
		return nil, fmt.Errorf("tencent poi failed: %s", result.Message)
	}

	attractions := make([]model.Attraction, 0)

	for _, item := range result.Data {
		if strings.TrimSpace(item.Title) == "" {
			continue
		}

		attraction := model.Attraction{
			Name:          item.Title,
			Category:      item.Category,
			Address:       item.Address,
			Description: buildPOIDescription(item.Title, item.Category, item.Address, keyword),
			EstimatedCost: 0,
			VisitTime:     "建议根据现场情况安排",
			Link:          buildTencentMapSearchLink(item.Title),
			Lat:           item.Location.Lat,
			Lng:           item.Location.Lng,
			DataSource:    "tencent_map",
		}

		attractions = append(attractions, attraction)
	}

	return attractions, nil
}

func buildPOIDescription(name string, category string, address string, keyword string) string {
	if category != "" && address != "" {
		return fmt.Sprintf("%s，搜索关键词：%s，类型：%s，地址：%s。", name, keyword, category, address)
	}

	if category != "" {
		return fmt.Sprintf("%s，搜索关键词：%s，类型：%s。", name, keyword, category)
	}

	if address != "" {
		return fmt.Sprintf("%s，搜索关键词：%s，地址：%s。", name, keyword, address)
	}

	return fmt.Sprintf("%s，搜索关键词：%s。", name, keyword)
}

func buildTencentMapSearchLink(keyword string) string {
	return "https://map.qq.com/?type=search&query=" + url.QueryEscape(keyword)
}

func buildPOIKeywords(request model.TripRequest) []string {
	switch request.Preference {
	case "美食":
		return []string{"美食", "小吃", "商圈", "夜市"}
	case "拍照":
		return []string{"地标", "景点", "公园", "步行街"}
	case "省钱":
		return []string{"公园", "免费景点", "街区", "博物馆"}
	case "轻松":
		return []string{"公园", "景点", "商圈", "博物馆"}
	default:
		return []string{"景点", "公园", "商圈"}
	}
}