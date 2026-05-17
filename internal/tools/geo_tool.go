package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"ariadne/internal/model"
)

type GeoTool struct {
	Name        string
	Description string
}

func NewGeoTool() GeoTool {
	return GeoTool{
		Name:        "geo_tool",
		Description: "使用腾讯位置服务将地址解析为经纬度",
	}
}

func (t GeoTool) Run(address string, mapConfig model.MapConfig) (model.Location, error) {
	return GeocodeWithTencent(address, mapConfig)
}

type tencentGeocodeResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"result"`
}

func GeocodeWithTencent(address string, mapConfig model.MapConfig) (model.Location, error) {
	if mapConfig.TencentMapKey == "" {
		return model.Location{}, errors.New("tencent map key is empty")
	}

	if address == "" {
		return model.Location{}, errors.New("address is empty")
	}

	endpoint := "https://apis.map.qq.com/ws/geocoder/v1/"
	requestURL := fmt.Sprintf(
		"%s?address=%s&key=%s&output=json",
		endpoint,
		url.QueryEscape(address),
		url.QueryEscape(mapConfig.TencentMapKey),
	)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(requestURL)
	if err != nil {
		return model.Location{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.Location{}, fmt.Errorf("tencent geocode http status: %d", resp.StatusCode)
	}

	var result tencentGeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.Location{}, err
	}

	if result.Status != 0 {
		return model.Location{}, fmt.Errorf("tencent geocode failed: %s", result.Message)
	}

	return model.Location{
		Address: address,
		Lat:     result.Result.Location.Lat,
		Lng:     result.Result.Location.Lng,
	}, nil
}