package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Handles high performance communication with Python Ai service
type EmbeddingClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// Response for the Python FastAPI endpoint
type EmbedRequest struct {
	Text string `json:"text"`
}

// Response from the Python FastAPI endpoint
type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// Initializes a client with optimized connection pooling and timeouts
func NewEmbeddingClient(baseURL string) *EmbeddingClient {
	return &EmbeddingClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, // Time to compute
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Sends text to the Python service and returns the vectorized result
func (c *EmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("cannot embed empty text")
	}
	reqBody, err := json.Marshal(EmbedRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	url := fmt.Sprintf("%s/embed", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Simple retry mechanism for network issues
	var resp *http.Response
	for i := 0; i < 3; i++ {
		resp, err = c.HTTPClient.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("embedding service unavailable after retries: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status: %d", resp.StatusCode)
	}
	var embedResp EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding result: %w", err)
	}
	return embedResp.Embedding, nil
}
