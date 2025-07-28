// Path: ./service/ai_service/prompts/prompts.go

package prompts

import _ "embed"

//go:embed chat.prompt
var ChatPrompt string

//go:embed summarize.prompt
var SummarizePrompt string