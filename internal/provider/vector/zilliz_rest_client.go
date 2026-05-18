package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ZillizRESTClient struct {
	Endpoint   string
	Token      string
	DBName     string
	HTTPClient *http.Client
}

type zillizResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func NewZillizRESTClient() ZillizRESTClient {
	endpoint := strings.TrimRight(os.Getenv("ZILLIZ_ENDPOINT"), "/")

	return ZillizRESTClient{
		Endpoint: endpoint,
		Token:    os.Getenv("ZILLIZ_TOKEN"),
		DBName:   os.Getenv("ZILLIZ_DB_NAME"), // 为空就不传
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c ZillizRESTClient) withDBName(body map[string]any) map[string]any {
	if c.DBName != "" {
		body["dbName"] = c.DBName
	}
	return body
}

func (c ZillizRESTClient) IsConfigured() bool {
	return c.Endpoint != "" && c.Token != ""
}

func (c ZillizRESTClient) doPost(ctx context.Context, path string, body any, out any) error {
	if !c.IsConfigured() {
		return errors.New("zilliz endpoint or token is empty")
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Request-Timeout", "20")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("zilliz http status %d: %s", resp.StatusCode, string(respBytes))
	}

	var base zillizResponse
	if err := json.Unmarshal(respBytes, &base); err != nil {
		return fmt.Errorf("decode zilliz response failed: %w; raw=%s", err, string(respBytes))
	}

	if base.Code != 0 {
		return fmt.Errorf("zilliz error code=%d message=%s", base.Code, base.Message)
	}

	if out != nil && len(base.Data) > 0 {
		if err := json.Unmarshal(base.Data, out); err != nil {
			return fmt.Errorf("decode zilliz data failed: %w; raw=%s", err, string(base.Data))
		}
	}

	return nil
}

func (c ZillizRESTClient) ListCollections(ctx context.Context) ([]string, error) {
	body := c.withDBName(map[string]any{})

	var data []string
	err := c.doPost(ctx, "/v2/vectordb/collections/list", body, &data)
	if err == nil {
		return data, nil
	}

	var objectData map[string]any
	err2 := c.doPost(ctx, "/v2/vectordb/collections/list", body, &objectData)
	if err2 == nil {
		return []string{}, nil
	}

	return nil, err
}

func (c ZillizRESTClient) CreateQuickCollection(ctx context.Context, collectionName string, dimension int) error {
	body := c.withDBName(map[string]any{
		"collectionName":    collectionName,
		"dimension":         dimension,
		"metricType":        "COSINE",
		"idType":            "VarChar",
		"autoID":            false,
		"primaryFieldName":  "id",
		"vectorFieldName":   "vector",
		"description":       "Ariadne vector memory collection",
		"params": map[string]any{
			"max_length":         512,
			"enableDynamicField": true,
		},
	})

	return c.doPost(ctx, "/v2/vectordb/collections/create", body, nil)
}

func (c ZillizRESTClient) Insert(ctx context.Context, collectionName string, data []map[string]any) error {
	body := map[string]any{
		"dbName":         c.DBName,
		"collectionName": collectionName,
		"data":           data,
	}

	return c.doPost(ctx, "/v2/vectordb/entities/insert", body, nil)
}

type SearchResult struct {
	ID       string         `json:"id"`
	Score    float64        `json:"distance"`
	Entity   map[string]any  `json:"entity"`
	Raw      map[string]any  `json:"-"`
}

func (c ZillizRESTClient) Search(
	ctx context.Context,
	collectionName string,
	vector []float32,
	limit int,
	filter string,
	outputFields []string,
) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 5
	}

	body := map[string]any{
		"dbName":         c.DBName,
		"collectionName": collectionName,
		"data":           [][]float32{vector},
		"annsField":      "vector",
		"limit":          limit,
		"outputFields":   outputFields,
	}

	if filter != "" {
		body["filter"] = filter
	}

	var data []map[string]any
	err := c.doPost(ctx, "/v2/vectordb/entities/search", body, &data)
	return data, err
}