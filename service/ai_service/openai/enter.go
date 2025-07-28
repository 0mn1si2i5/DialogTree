// Path: ./service/ai_service/openai/enter.go

package openai

import (
	"dialogTree/global"
	"dialogTree/service/ai_service/common"
)

const baseURL = "https://api.openai.com/v1/chat/completions"

type ModelType string

const (
	GPT_35Turbo      ModelType = "gpt-3.5-turbo"
	GPT_4            ModelType = "gpt-4"
	GPT_4O           ModelType = "gpt-4o"
	GPT_4OMini       ModelType = "gpt-4o-mini"
	GPT_4Turbo       ModelType = "gpt-4-turbo"
	GPT_O1           ModelType = "o1"
	GPT_O1Mini       ModelType = "o1-mini"
	GPT_O1Preview    ModelType = "o1-preview"
)

// getConfig 获取OpenAI配置
func getConfig() common.AIProviderConfig {
	return common.AIProviderConfig{
		BaseURL: baseURL,
		APIKey:  global.Config.Ai.OpenAI.SecretKey,
		Model:   global.Config.Ai.OpenAI.Model,
	}
}