// Path: ./service/vector_service/common/common.go

package common

type SearchResult struct {
	ID       uint64                 `json:"id"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
	Vector   []float32              `json:"vector"`
}
