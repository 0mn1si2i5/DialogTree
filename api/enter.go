// Path: ./api/enter.go

package api

import (
	"dialogTree/api/category_api"
	"dialogTree/api/dialog_api"
	"dialogTree/api/session_api"
)

type Api struct {
	SessionApi  session_api.SessionApi
	DialogApi   dialog_api.DialogApi
	CategoryApi category_api.CategoryApi
}

var App = new(Api)
