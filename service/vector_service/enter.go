// Path: ./service/vector_service/enter.go

package vector_service

import (
	"dialogTree/service/vector_service/common"
	"dialogTree/service/vector_service/qdrant_service"
)

type VectorService interface {
	// 存储向量和元数据
	Store(id string, vector []float32, metadata map[string]interface{}) error
	
	// 向量检索
	Search(vector []float32, topK int, filter map[string]interface{}) ([]common.SearchResult, error)
	
	// 删除向量
	Delete(id string) error
	
	// 初始化集合
	InitCollection() error
}

var VectorServiceInstance VectorService

func InitVectorService() error {
	// 根据配置选择向量数据库实现
	VectorServiceInstance = &qdrant_service.QdrantService{}
	return VectorServiceInstance.InitCollection()
}
