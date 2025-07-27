// Path: ./service/vector_service/qdrant_service/qdrant.go

package qdrant_service

import (
	"bytes"
	"dialogTree/global"
	"dialogTree/service/vector_service/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type QdrantService struct {
	baseURL    string
	collection string
	client     *http.Client
}

type QdrantPoint struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

type QdrantSearchRequest struct {
	Vector      []float32              `json:"vector"`
	Limit       int                    `json:"limit"`
	WithPayload bool                   `json:"with_payload"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
}

type QdrantSearchResponse struct {
	Result []struct {
		ID      string                 `json:"id"`
		Score   float64                `json:"score"`
		Payload map[string]interface{} `json:"payload"`
	} `json:"result"`
}

func (q *QdrantService) init() {
	config := global.Config.Vector.Qdrant
	q.baseURL = fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	q.collection = config.Collection
	q.client = &http.Client{Timeout: 30 * time.Second}
}

func (q *QdrantService) InitCollection() error {
	q.init()
	
	// 创建集合的配置
	createReq := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     1536, // text-embedding-3-small 的维度
			"distance": "Cosine",
		},
	}
	
	reqBody, _ := json.Marshal(createReq)
	url := fmt.Sprintf("%s/collections/%s", q.baseURL, q.collection)
	
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	if global.Config.Vector.Qdrant.ApiKey != "" {
		req.Header.Set("Api-Key", global.Config.Vector.Qdrant.ApiKey)
	}
	
	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// 409 表示集合已存在，这是正常的
	if resp.StatusCode != 200 && resp.StatusCode != 409 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: %d, %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func (q *QdrantService) Store(id string, vector []float32, metadata map[string]interface{}) error {
	point := QdrantPoint{
		ID:      id,
		Vector:  vector,
		Payload: metadata,
	}
	
	reqBody := map[string]interface{}{
		"points": []QdrantPoint{point},
	}
	
	return q.makeRequest("PUT", fmt.Sprintf("/collections/%s/points", q.collection), reqBody)
}

func (q *QdrantService) Search(vector []float32, topK int, filter map[string]interface{}) ([]common.SearchResult, error) {
	searchReq := QdrantSearchRequest{
		Vector:      vector,
		Limit:       topK,
		WithPayload: true,
		Filter:      filter,
	}

	reqBody, _ := json.Marshal(searchReq)
	url := fmt.Sprintf("%s/collections/%s/points/search", q.baseURL, q.collection)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if global.Config.Vector.Qdrant.ApiKey != "" {
		req.Header.Set("Api-Key", global.Config.Vector.Qdrant.ApiKey)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var qdrantResp QdrantSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&qdrantResp); err != nil {
		return nil, err
	}

	results := make([]common.SearchResult, 0, len(qdrantResp.Result))
	for _, item := range qdrantResp.Result {
		if item.Score >= global.Config.Vector.SimilarityThreshold {
			results = append(results, common.SearchResult{
				ID:       item.ID,
				Score:    item.Score,
				Metadata: item.Payload,
			})
		}
	}

	return results, nil
}

func (q *QdrantService) Delete(id string) error {
	reqBody := map[string]interface{}{
		"points": []string{id},
	}
	
	return q.makeRequest("POST", fmt.Sprintf("/collections/%s/points/delete", q.collection), reqBody)
}

func (q *QdrantService) makeRequest(method, path string, body interface{}) error {
	reqBody, _ := json.Marshal(body)
	url := q.baseURL + path
	
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	if global.Config.Vector.Qdrant.ApiKey != "" {
		req.Header.Set("Api-Key", global.Config.Vector.Qdrant.ApiKey)
	}
	
	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %d, %s", resp.StatusCode, string(body))
	}
	
	return nil
}