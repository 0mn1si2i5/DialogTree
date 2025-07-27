// Path: ./service/embedding_service/embedding.go

package embedding_service

import (
	"bytes"
	"dialogTree/global"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

type EmbeddingService struct {
	client *http.Client
}

var EmbeddingServiceInstance *EmbeddingService

func InitEmbeddingService() {
	EmbeddingServiceInstance = &EmbeddingService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *EmbeddingService) GetEmbedding(text string) ([]float32, error) {
	reqBody := EmbeddingRequest{
		Input: text,
		Model: global.Config.Ai.EmbeddingModel,
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	
	// 使用 ChatAnywhere 的配置（通常支持 embedding API）
	req, err := http.NewRequest("POST", "https://api.chatanywhere.tech/v1/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+global.Config.Ai.ChatAnywhere.SecretKey)
	
	resp, err := e.client.Do(req)
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

func GetEmbedding(text string) ([]float32, error) {
	return EmbeddingServiceInstance.GetEmbedding(text)
}