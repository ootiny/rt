package _rt_package_name_

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var gAPIMap = map[string]func(ctx *Context, data []byte) *Return{}

const ErrInternal = 1001
const ErrActionNotFound = 1002
const ErrActionNotImplemented = 1003
const ErrMethodNotSupported = 1004
const ErrReadData = 1005
const ErrUnmarshalData = 1006
const ErrAPIExec = 2001

type Error interface {
	GetCode() int
	Error() string
}

type GoError struct {
	code int
	err  error
}

func NewError(code int, err error) Error {
	return &GoError{
		code: code,
		err:  err,
	}
}

func (p *GoError) GetCode() int {
	return p.code
}
func (p *GoError) Error() string {
	return p.err.Error()
}

type Request interface {
	Action() string
	Data() []byte
	Cookie(name string) (*http.Cookie, error)
	Header(name string) string
}

type Response interface {
	SetHeader(name string, value string)
	WriteHeader(code int)
	WriteJson(data []byte) (int, error)
}

func NewContext(request Request, response Response) *Context {
	return &Context{
		request:  request,
		response: response,
	}
}

type Context struct {
	request  Request
	response Response
}

func (p *Context) Request() Request {
	return p.request
}

func (p *Context) Response() Response {
	return p.response
}

func (p *Context) Errorf(format string, args ...any) Error {
	return NewError(ErrInternal, fmt.Errorf(format, args...))
}

func (p *Context) ErrorWithCodef(code int, format string, args ...any) Error {
	return NewError(code, fmt.Errorf(format, args...))
}

type Return struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func RegisterHandler(action string, handler func(ctx *Context, data []byte) *Return) {
	gAPIMap[action] = handler
}

func JsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func apiHandler(cors bool, w Response, r Request) {
	// Set CORS headers
	if cors {
		w.SetHeader("Access-Control-Allow-Origin", "*")

		// Handle preflight OPTIONS request only when CORS is enabled
		if r.Header("Access-Control-Request-Method") == "OPTIONS" {
			w.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	action := r.Action()
	ret := (*Return)(nil)
	fn, ok := gAPIMap[action]

	if !ok {
		ret = &Return{
			Code:    ErrActionNotFound,
			Message: fmt.Sprintf("action %s not found", action),
		}
	} else {
		data := r.Data()
		ctx := NewContext(r, w)
		ret = fn(ctx, data)
	}

	if ret.Code == 0 {
		ret.Code = http.StatusOK
	}

	w.SetHeader("Content-Type", "application/json")

	if retBytes, err := json.Marshal(ret); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.WriteJson([]byte(`{"code":500,"message":"Internal Server Error"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = w.WriteJson(retBytes)
	}
}
