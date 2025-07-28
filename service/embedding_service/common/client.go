// Path: ./service/embedding_service/common/client.go

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingRequest 通用的embedding请求结构
type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// EmbeddingResponse 通用的embedding响应结构
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// EmbeddingProviderConfig embedding提供商配置
type EmbeddingProviderConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

// MakeEmbeddingRequest 通用的embedding HTTP请求函数
func MakeEmbeddingRequest(config EmbeddingProviderConfig, text string) ([]float32, error) {
	reqBody := EmbeddingRequest{
		Input: text,
		Model: config.Model,
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", config.BaseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding request failed: %d, %s", resp.StatusCode, string(body))
	}
	
	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, err
	}
	
	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}
	
	return embeddingResp.Data[0].Embedding, nil
}