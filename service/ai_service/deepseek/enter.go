// Path: ./service/ai_service/deepseek/enter.go

package deepseek

import (
	"dialogTree/global"
	"dialogTree/service/ai_service/common"
)

const baseURL = "https://api.deepseek.com/v1/chat/completions"

type ModelType string

const (
	DeepSeekChat      ModelType = "deepseek-chat"
	DeepSeekCoder     ModelType = "deepseek-coder"
	DeepSeekV3        ModelType = "deepseek-v3"
	DeepSeekR1        ModelType = "deepseek-r1"
	DeepSeekR1Distill ModelType = "deepseek-r1-distill-llama-70b"
)

// getConfig 获取DeepSeek配置
func getConfig() common.AIProviderConfig {
	return common.AIProviderConfig{
		BaseURL: baseURL,
		APIKey:  global.Config.Ai.DeepSeek.SecretKey,
		Model:   global.Config.Ai.DeepSeek.Model,
	}
}