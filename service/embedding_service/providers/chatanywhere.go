// Path: ./service/embedding_service/providers/chatanywhere.go

package providers

import (
	"dialogTree/global"
	"dialogTree/service/embedding_service/common"
	"fmt"
)

const chatanywhereEmbeddingURL = "https://api.chatanywhere.tech/v1/embeddings"

// ChatAnywhereEmbedding 获取ChatAnywhere embedding
func ChatAnywhereEmbedding(text string) ([]float32, error) {
	if global.Config.Ai.ChatAnywhere.SecretKey == "" {
		return nil, fmt.Errorf("ChatAnywhere API密钥未配置")
	}
	
	config := common.EmbeddingProviderConfig{
		BaseURL: chatanywhereEmbeddingURL,
		APIKey:  global.Config.Ai.ChatAnywhere.SecretKey,
		Model:   global.Config.Ai.EmbeddingModel,
	}
	
	return common.MakeEmbeddingRequest(config, text)
}