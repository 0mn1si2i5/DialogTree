// Path: ./service/ai_service/chat_anywhere_ai/enter.go

package chat_anywhere_ai

import (
	"dialogTree/global"
	"dialogTree/service/ai_service/common"
)

const baseURL = "https://api.chatanywhere.tech/v1/chat/completions"

type ModelType string

const (
	CA_GPT_35Turbo ModelType = "gpt-3.5-turbo"
	CA_GPT_O1Mini  ModelType = "o1-mini"
	CA_GPT_4O      ModelType = "gpt-4o"
	CA_DS_R1       ModelType = "deepseek-r1"
	CA_DS_V3       ModelType = "deepseek-v3"
	CA_CLA_O4      ModelType = "claude-opus-4-20250514"
	CA_CLA_S4      ModelType = "claude-sonnet-4-20250514"
)

// getConfig 获取ChatAnywhere配置
func getConfig() common.AIProviderConfig {
	return common.AIProviderConfig{
		BaseURL: baseURL,
		APIKey:  global.Config.Ai.ChatAnywhere.SecretKey,
		Model:   global.Config.Ai.ChatAnywhere.Model,
	}
}
