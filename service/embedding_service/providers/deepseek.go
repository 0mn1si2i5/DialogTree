// Path: ./service/embedding_service/providers/deepseek.go

package providers

import (
	"dialogTree/global"
	"dialogTree/service/embedding_service/common"
	"fmt"
)

const deepseekEmbeddingURL = "https://api.deepseek.com/v1/embeddings"

// DeepSeekEmbedding 获取DeepSeek embedding
func DeepSeekEmbedding(text string) ([]float32, error) {
	if global.Config.Ai.DeepSeek.SecretKey == "" {
		return nil, fmt.Errorf("DeepSeek API密钥未配置")
	}
	
	config := common.EmbeddingProviderConfig{
		BaseURL: deepseekEmbeddingURL,
		APIKey:  global.Config.Ai.DeepSeek.SecretKey,
		Model:   global.Config.Ai.EmbeddingModel,
	}
	
	return common.MakeEmbeddingRequest(config, text)
}