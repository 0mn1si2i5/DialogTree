// Path: ./service/vector_service/enter.go

package vector_service

import (
	"dialogTree/service/vector_service/common"
	"dialogTree/service/vector_service/qdrant_service"
)

type VectorService interface {
	// 存储向量和元数据
	Store(id uint64, vector []float32, metadata map[string]interface{}) error
	
	// 向量检索
	Search(vector []float32, topK int, filter map[string]interface{}) ([]common.SearchResult, error)
	
	// 删除向量
	Delete(id uint64) error
	
	// 初始化集合
	InitCollection() error

	// 获取所有点（用于调试和验证）
	GetAllPoints() ([]common.SearchResult, error)

	// 清空集合
	ClearCollection() error
}

var VectorServiceInstance VectorService

func InitVectorService() error {
	// 根据配置选择向量数据库实现
	VectorServiceInstance = &qdrant_service.QdrantService{}
	return VectorServiceInstance.InitCollection()
}
