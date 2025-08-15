package _rt_package_name_

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var gHttpActionMap = map[string]func(ctx Context, w Response, data []byte) *Return{}

const ErrInternal = 1001
const ErrActionNotFound = 1002
const ErrActionNotImplemented = 1003
const ErrMethodNotSupported = 1004
const ErrReadData = 1005
const ErrUnmarshalData = 1006

type Error interface {
	GetCode() int
	Error() string
}

type Response interface {
	WriteJson(data []byte) (int, error)
}

type RequestContext interface {
	Action() string
	Cookie(name string) (*http.Cookie, error)
	Header(name string) string
}

type Context interface {
	Request() RequestContext
	Errorf(format string, args ...any) Error
	ErrorWithCodef(code int, format string, args ...any) Error
}

type Return struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
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

type ErrorContext struct{}

func (p *ErrorContext) Errorf(format string, args ...any) Error {
	return NewError(ErrInternal, fmt.Errorf(format, args...))
}

func (p *ErrorContext) ErrorWithCodef(code int, format string, args ...any) Error {
	return NewError(code, fmt.Errorf(format, args...))
}

func RegisterHandler(action string, handler func(ctx Context, response Response, data []byte) *Return) {
	gHttpActionMap[action] = handler
}

func JsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func CallAction(url string, action string, method string, data []byte) ([]byte, error) {
	switch method {
	case http.MethodGet:
		return nil, fmt.Errorf("method %s not supported", method)
	case http.MethodPost:
		return nil, fmt.Errorf("method %s not supported", method)
	default:
		return nil, fmt.Errorf("method %s not supported", method)
	}
}
