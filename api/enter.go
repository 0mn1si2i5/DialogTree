// Path: ./api/enter.go

package api

import "dialogTree/api/ai_api"

type Api struct {
	AiApi ai_api.AiApi
}

var App = new(Api)
