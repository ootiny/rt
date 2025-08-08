package _rt_system_

import "net/http"

var gHttpActionMap = map[string]func(ctx IContext, w IResponse, data []byte) *Return{}

const ErrAPIActionNotFound = 1001
const ErrAPIReadData = 1002

type IResponse interface {
	WriteJson(data []byte) (int, error)
}

type IContext interface {
	Action() string
	Method() string
	Cookie(name string) (*http.Cookie, error)
	Header(name string) string
}

type IServer interface {
	Run() error
}

type Return struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func RegisterHandler(action string, handler func(ctx IContext, response IResponse, data []byte) *Return) {
	gHttpActionMap[action] = handler
}
