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

type DistanceMatrixTool struct {
	Name        string
	Description string
}

func NewDistanceMatrixTool() DistanceMatrixTool {
	return DistanceMatrixTool{
		Name:        "distance_matrix_tool",
		Description: "使用腾讯位置服务距离矩阵 API 计算每日景点之间的真实路面距离和预计时间",
	}
}

func (t DistanceMatrixTool) Run(routes []model.DailyRoute, mapConfig model.MapConfig) []model.DailyRoute {
	for i := range routes {
		routes[i].RouteSegments = buildRouteSegments(routes[i].Attractions, mapConfig)
	}

	return routes
}

func buildRouteSegments(attractions []model.Attraction, mapConfig model.MapConfig) []model.RouteSegment {
	segments := make([]model.RouteSegment, 0)

	if len(attractions) < 2 {
		return segments
	}

	for i := 0; i < len(attractions)-1; i++ {
		from := attractions[i]
		to := attractions[i+1]

		segment := queryRouteSegmentWithTencent(from, to, mapConfig)
		segments = append(segments, segment)
	}

	return segments
}

type tencentDistanceMatrixResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Rows []struct {
			Elements []struct {
				Distance int `json:"distance"`
				Duration int `json:"duration"`
			} `json:"elements"`
		} `json:"rows"`
	} `json:"result"`
}

func queryRouteSegmentWithTencent(from model.Attraction, to model.Attraction, mapConfig model.MapConfig) model.RouteSegment {
	segment := model.RouteSegment{
		From:       from.Name,
		To:         to.Name,
		Mode:       "driving",
		DataSource: "tencent_map",
		Status:     "unavailable",
	}

	distance, durationSeconds, err := getDistanceMatrixWithTencent(from, to, mapConfig)
	if err != nil {
		segment.Message = err.Error()
		return segment
	}

	segment.DistanceMeters = distance
	segment.DurationSeconds = durationSeconds
	segment.DurationMinutes = secondsToMinutes(durationSeconds)
	segment.Status = "ok"
	segment.Message = "query ok"

	return segment
}

func getDistanceMatrixWithTencent(from model.Attraction, to model.Attraction, mapConfig model.MapConfig) (int, int, error) {
	if mapConfig.TencentMapKey == "" {
		return 0, 0, errors.New("tencent map key is empty")
	}

	if from.Lat == 0 || from.Lng == 0 {
		return 0, 0, errors.New("from attraction location is empty")
	}

	if to.Lat == 0 || to.Lng == 0 {
		return 0, 0, errors.New("to attraction location is empty")
	}

	endpoint := "https://apis.map.qq.com/ws/distance/v1/matrix/"

	query := url.Values{}
	query.Set("mode", "driving")
	query.Set("from", fmt.Sprintf("%f,%f", from.Lat, from.Lng))
	query.Set("to", fmt.Sprintf("%f,%f", to.Lat, to.Lng))
	query.Set("output", "json")
	query.Set("key", mapConfig.TencentMapKey)

	requestURL := endpoint + "?" + query.Encode()

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	resp, err := client.Get(requestURL)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("tencent distance matrix http status: %d", resp.StatusCode)
	}

	var result tencentDistanceMatrixResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}

	if result.Status != 0 {
		return 0, 0, fmt.Errorf("tencent distance matrix failed: %s", result.Message)
	}

	if len(result.Result.Rows) == 0 || len(result.Result.Rows[0].Elements) == 0 {
		return 0, 0, errors.New("tencent distance matrix returned empty elements")
	}

	element := result.Result.Rows[0].Elements[0]

	return element.Distance, element.Duration, nil
}

func secondsToMinutes(seconds int) int {
	if seconds <= 0 {
		return 0
	}

	return (seconds + 59) / 60
}