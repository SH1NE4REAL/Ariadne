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
	attractions, err := SearchAttractionsWithTencent(request.Destination, mapConfig)
	if err != nil {
		return []model.Attraction{}
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

func SearchAttractionsWithTencent(destination string, mapConfig model.MapConfig) ([]model.Attraction, error) {
	if mapConfig.TencentMapKey == "" {
		return nil, errors.New("tencent map key is empty")
	}

	if destination == "" {
		return nil, errors.New("destination is empty")
	}

	endpoint := "https://apis.map.qq.com/ws/place/v1/search"

	query := url.Values{}
	query.Set("boundary", fmt.Sprintf("region(%s,0)", destination))
	query.Set("keyword", "景点")
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
			Description:   buildPOIDescription(item.Title, item.Category, item.Address),
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

func buildPOIDescription(name string, category string, address string) string {
	if category != "" && address != "" {
		return fmt.Sprintf("%s，类型：%s，地址：%s。", name, category, address)
	}

	if category != "" {
		return fmt.Sprintf("%s，类型：%s。", name, category)
	}

	if address != "" {
		return fmt.Sprintf("%s，地址：%s。", name, address)
	}

	return name
}

func buildTencentMapSearchLink(keyword string) string {
	return "https://map.qq.com/?type=search&query=" + url.QueryEscape(keyword)
}