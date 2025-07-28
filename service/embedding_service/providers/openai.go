// Path: ./service/embedding_service/providers/openai.go

package providers

import (
	"dialogTree/global"
	"dialogTree/service/embedding_service/common"
	"fmt"
)

const openaiEmbeddingURL = "https://api.openai.com/v1/embeddings"

// OpenAIEmbedding 获取OpenAI embedding
func OpenAIEmbedding(text string) ([]float32, error) {
	if global.Config.Ai.OpenAI.SecretKey == "" {
		return nil, fmt.Errorf("OpenAI API密钥未配置")
	}
	
	config := common.EmbeddingProviderConfig{
		BaseURL: openaiEmbeddingURL,
		APIKey:  global.Config.Ai.OpenAI.SecretKey,
		Model:   global.Config.Ai.EmbeddingModel,
	}
	
	return common.MakeEmbeddingRequest(config, text)
}