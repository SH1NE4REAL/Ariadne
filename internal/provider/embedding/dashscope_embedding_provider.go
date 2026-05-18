package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type DashScopeEmbeddingProvider struct {
	APIKey     string
	BaseURL    string
	Model      string
	Dimension  int
	HTTPClient *http.Client
}

type dashScopeEmbeddingRequest struct {
	Model      string                  `json:"model"`
	Input      dashScopeEmbeddingInput  `json:"input"`
	Parameters dashScopeEmbeddingParams `json:"parameters"`
}

type dashScopeEmbeddingInput struct {
	Texts []string `json:"texts"`
}

type dashScopeEmbeddingParams struct {
	Dimension  int    `json:"dimension"`
	TextType   string `json:"text_type,omitempty"`   // query / document
	OutputType string `json:"output_type,omitempty"` // dense
}

type dashScopeEmbeddingResponse struct {
	StatusCode int    `json:"status_code"`
	RequestID  string `json:"request_id"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Output     struct {
		Embeddings []struct {
			Embedding []float64 `json:"embedding"`
			TextIndex int       `json:"text_index"`
		} `json:"embeddings"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func NewDashScopeEmbeddingProvider() DashScopeEmbeddingProvider {
	baseURL := strings.TrimRight(os.Getenv("DASHSCOPE_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/api/v1"
	}

	model := os.Getenv("EMBEDDING_MODEL")
	if model == "" {
		model = "text-embedding-v4"
	}

	dimension := 1024
	if text := os.Getenv("EMBEDDING_DIM"); text != "" {
		value, err := strconv.Atoi(text)
		if err == nil && value > 0 {
			dimension = value
		}
	}

	return DashScopeEmbeddingProvider{
		APIKey:    os.Getenv("DASHSCOPE_API_KEY"),
		BaseURL:   baseURL,
		Model:     model,
		Dimension: dimension,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p DashScopeEmbeddingProvider) IsConfigured() bool {
	return p.APIKey != ""
}

func (p DashScopeEmbeddingProvider) EmbedDocument(ctx context.Context, text string) ([]float32, error) {
	return p.embed(ctx, text, "document")
}

func (p DashScopeEmbeddingProvider) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return p.embed(ctx, text, "query")
}

func (p DashScopeEmbeddingProvider) embed(ctx context.Context, text string, textType string) ([]float32, error) {
	if !p.IsConfigured() {
		return nil, errors.New("DASHSCOPE_API_KEY is empty")
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("embedding text is empty")
	}

	reqBody := dashScopeEmbeddingRequest{
		Model: p.Model,
		Input: dashScopeEmbeddingInput{
			Texts: []string{text},
		},
		Parameters: dashScopeEmbeddingParams{
			Dimension:  p.Dimension,
			TextType:   textType,
			OutputType: "dense",
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := p.BaseURL + "/services/embeddings/text-embedding/text-embedding"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("dashscope http status %d: %s", resp.StatusCode, string(respBytes))
	}

	var result dashScopeEmbeddingResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("decode dashscope response failed: %w; raw=%s", err, string(respBytes))
	}

	if result.Code != "" {
		return nil, fmt.Errorf("dashscope error code=%s message=%s", result.Code, result.Message)
	}

	if len(result.Output.Embeddings) == 0 {
		return nil, errors.New("dashscope returned empty embedding")
	}

	embedding64 := result.Output.Embeddings[0].Embedding
	if len(embedding64) == 0 {
		return nil, errors.New("dashscope embedding vector is empty")
	}

	embedding32 := make([]float32, len(embedding64))
	for i, value := range embedding64 {
		embedding32[i] = float32(value)
	}

	if len(embedding32) != p.Dimension {
		return nil, fmt.Errorf("embedding dimension mismatch: got %d, want %d", len(embedding32), p.Dimension)
	}

	return embedding32, nil
}