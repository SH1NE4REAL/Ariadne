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

type RouteDistanceTool struct {
	Name        string
	Description string
}

func NewRouteDistanceTool() RouteDistanceTool {
	return RouteDistanceTool{
		Name:        "route_distance_tool",
		Description: "使用腾讯位置服务路线规划 API 获取真实驾车距离和预计时间",
	}
}

func (t RouteDistanceTool) Run(
	origin string,
	destination string,
	originLocation model.Location,
	destinationLocation model.Location,
	mapConfig model.MapConfig,
) model.RouteDistance {
	result, err := GetDrivingRouteDistanceWithTencent(
		origin,
		destination,
		originLocation,
		destinationLocation,
		mapConfig,
	)

	if err != nil {
		return model.RouteDistance{
			Origin:      origin,
			Destination: destination,
			Mode:        "driving",
			DataSource:  "tencent_map",
			Status:      "unavailable",
			Message:     err.Error(),
		}
	}

	return result
}

type tencentDirectionResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Routes []struct {
			Mode     string `json:"mode"`
			Distance int    `json:"distance"`
			Duration int    `json:"duration"`
			TaxiFare struct {
				Fare int `json:"fare"`
			} `json:"taxi_fare"`
		} `json:"routes"`
	} `json:"result"`
}

func GetDrivingRouteDistanceWithTencent(
	origin string,
	destination string,
	originLocation model.Location,
	destinationLocation model.Location,
	mapConfig model.MapConfig,
) (model.RouteDistance, error) {
	if mapConfig.TencentMapKey == "" {
		return model.RouteDistance{}, errors.New("tencent map key is empty")
	}

	if originLocation.Lat == 0 || originLocation.Lng == 0 {
		return model.RouteDistance{}, errors.New("origin location is empty")
	}

	if destinationLocation.Lat == 0 || destinationLocation.Lng == 0 {
		return model.RouteDistance{}, errors.New("destination location is empty")
	}

	endpoint := "https://apis.map.qq.com/ws/direction/v1/driving/"

	requestURL := fmt.Sprintf(
		"%s?from=%f,%f&to=%f,%f&policy=LEAST_TIME&output=json&key=%s",
		endpoint,
		originLocation.Lat,
		originLocation.Lng,
		destinationLocation.Lat,
		destinationLocation.Lng,
		url.QueryEscape(mapConfig.TencentMapKey),
	)

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	resp, err := client.Get(requestURL)
	if err != nil {
		return model.RouteDistance{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.RouteDistance{}, fmt.Errorf("tencent direction http status: %d", resp.StatusCode)
	}

	var result tencentDirectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.RouteDistance{}, err
	}

	if result.Status != 0 {
		return model.RouteDistance{}, fmt.Errorf("tencent direction failed: %s", result.Message)
	}

	if len(result.Result.Routes) == 0 {
		return model.RouteDistance{}, errors.New("tencent direction returned empty routes")
	}

	route := result.Result.Routes[0]

	return model.RouteDistance{
		Origin:            origin,
		Destination:       destination,
		Mode:              route.Mode,
		DistanceMeters:    route.Distance,
		DurationMinutes:   route.Duration,
		EstimatedTaxiFare: route.TaxiFare.Fare,
		DataSource:        "tencent_map",
		Status:            "ok",
		Message:           "query ok",
	}, nil
}