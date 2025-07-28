// Path: ./service/embedding_service/embedding.go

package embedding_service

import (
	"dialogTree/global"
	"dialogTree/service/embedding_service/providers"
	"fmt"
	"strings"
)

type EmbeddingProvider string

const (
	OpenAIProvider       EmbeddingProvider = "openai"
	DeepSeekProvider     EmbeddingProvider = "deepseek"
	ChatAnywhereProvider EmbeddingProvider = "chatanywhere"
)

type EmbeddingService struct{}

var EmbeddingServiceInstance *EmbeddingService

func InitEmbeddingService() {
	EmbeddingServiceInstance = &EmbeddingService{}
}

// GetEmbedding 获取embedding，根据配置的provider选择对应的服务
func (e *EmbeddingService) GetEmbedding(text string) ([]float32, error) {
	provider := EmbeddingProvider(strings.ToLower(global.Config.Ai.EmbeddingProvider))
	
	switch provider {
	case OpenAIProvider:
		return providers.OpenAIEmbedding(text)
	case DeepSeekProvider:
		return providers.DeepSeekEmbedding(text)
	case ChatAnywhereProvider:
		return providers.ChatAnywhereEmbedding(text)
	default:
		// 如果没有配置或配置错误，尝试自动选择可用的提供商
		return e.getEmbeddingWithFallback(text)
	}
}

// getEmbeddingWithFallback 自动选择可用的embedding提供商
func (e *EmbeddingService) getEmbeddingWithFallback(text string) ([]float32, error) {
	// 优先级：OpenAI > DeepSeek > ChatAnywhere
	var lastErr error
	
	// 尝试OpenAI
	if global.Config.Ai.OpenAI.SecretKey != "" {
		result, err := providers.OpenAIEmbedding(text)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	
	// 尝试DeepSeek
	if global.Config.Ai.DeepSeek.SecretKey != "" {
		result, err := providers.DeepSeekEmbedding(text)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	
	// 尝试ChatAnywhere
	if global.Config.Ai.ChatAnywhere.SecretKey != "" {
		result, err := providers.ChatAnywhereEmbedding(text)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	
	// 所有提供商都不可用
	if lastErr != nil {
		return nil, fmt.Errorf("所有embedding提供商都不可用，最后错误: %v", lastErr)
	}
	return nil, fmt.Errorf("没有配置可用的embedding提供商API密钥")
}

// GetEmbedding 全局函数，保持向后兼容
func GetEmbedding(text string) ([]float32, error) {
	return EmbeddingServiceInstance.GetEmbedding(text)
}