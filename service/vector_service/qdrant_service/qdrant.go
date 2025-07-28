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
	ID      uint64                 `json:"id"`
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
		ID      uint64                 `json:"id"`
		Score   float64                `json:"score"`
		Payload map[string]interface{} `json:"payload"`
	} `json:"result"`
}

// QdrantScrollRequest 用于Qdrant的scroll API请求
type QdrantScrollRequest struct {
	Limit       *int                   `json:"limit,omitempty"`
	Offset      *uint64                `json:"offset,omitempty"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
	WithPayload bool                   `json:"with_payload"`
	WithVectors bool                   `json:"with_vectors"`
}

// QdrantScrollResponse 用于Qdrant的scroll API响应
type QdrantScrollResponse struct {
	Result struct {
		Points []struct {
			ID      uint64                 `json:"id"`
			Payload map[string]interface{} `json:"payload"`
			Vector  []float32              `json:"vector"`
		} `json:"points"`
		NextPageOffset *uint64 `json:"next_page_offset"`
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

func (q *QdrantService) Store(id uint64, vector []float32, metadata map[string]interface{}) error {
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

func (q *QdrantService) Delete(id uint64) error {
	reqBody := map[string]interface{}{
		"points": []uint64{id},
	}

	return q.makeRequest("POST", fmt.Sprintf("/collections/%s/points/delete", q.collection), reqBody)
}

func (q *QdrantService) GetAllPoints() ([]common.SearchResult, error) {
	var allResults []common.SearchResult
	limit := 100             // 每页获取100个点
	var offset *uint64 = nil // 初始偏移量为nil

	for {
		scrollReq := QdrantScrollRequest{
			Limit:       &limit,
			Offset:      offset,
			WithPayload: true,
			WithVectors: true, // 通常不需要向量本身，只需要ID和元数据
		}

		respBody, err := q.makeRequestWithResponse("POST", fmt.Sprintf("/collections/%s/points/scroll", q.collection), scrollReq)
		if err != nil {
			return nil, fmt.Errorf("scroll request failed: %v", err)
		}

		var scrollResp QdrantScrollResponse
		if err := json.Unmarshal(respBody, &scrollResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scroll response: %v", err)
		}

		for _, point := range scrollResp.Result.Points {
			allResults = append(allResults, common.SearchResult{
				ID:       point.ID,
				Metadata: point.Payload,
				// Score is not available in scroll results, set to 0 or omit
				Score:  0,
				Vector: point.Vector,
			})
		}

		if scrollResp.Result.NextPageOffset == nil {
			// 没有更多点，退出循环
			break
		}
		offset = scrollResp.Result.NextPageOffset
	}

	return allResults, nil
}

func (q *QdrantService) makeRequest(method, path string, body interface{}) error {
	_, err := q.makeRequestWithResponse(method, path, body)
	return err
}

func (q *QdrantService) makeRequestWithResponse(method, path string, body interface{}) ([]byte, error) {
	reqBody, _ := json.Marshal(body)
	url := q.baseURL + path

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed: %d, %s", resp.StatusCode, string(body))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return respBody, nil
}

func (q *QdrantService) ClearCollection() error {
	url := fmt.Sprintf("/collections/%s", q.collection)
	return q.makeRequest("DELETE", url, nil)
}
