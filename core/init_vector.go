// Path: ./core/init_vector.go

package core

import (
	"dialogTree/global"
	"dialogTree/service/embedding_service"
	"dialogTree/service/vector_service"
	"fmt"
)

func InitVector() error {
	if global.Config.Vector.Enable == false {
		return nil
	}
	// 初始化向量数据库服务
	err := vector_service.InitVectorService()
	if err != nil {
		return fmt.Errorf("初始化向量数据库失败: %v", err)
	}

	// 初始化 embedding 服务
	embedding_service.InitEmbeddingService()

	fmt.Println("向量服务初始化完成")
	return nil
}
