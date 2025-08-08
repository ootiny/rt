package _rt_system_

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var gMap = map[string]func(ctx *Context) *Return{}

const ErrAPIActionNotFound = 1001
const ErrAPIReadData = 1002

type Return struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Context struct {
	action string
	data   []byte
	w      http.ResponseWriter
	r      *http.Request
}

func (p *Context) Action() string {
	return p.action
}

func (p *Context) Data() []byte {
	return p.data
}

func (p *Context) Method() string {
	return p.r.Method
}

func (p *Context) Cookie(name string) (*http.Cookie, error) {
	return p.r.Cookie(name)
}

func (p *Context) Header(name string) string {
	return p.r.Header.Get(name)
}

func (p *Context) WriteJson(data []byte) (int, error) {
	p.w.Header().Set("Content-Type", "application/json")
	return p.w.Write(data)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("a")
	ret := (*Return)(nil)
	fn, ok := gMap[action]

	if !ok {
		ret = &Return{
			Code:    ErrAPIActionNotFound,
			Message: fmt.Sprintf("action %s not found", action),
		}
	} else {
		data := []byte(r.URL.Query().Get("d"))

		if len(data) == 0 {
			if body, err := io.ReadAll(r.Body); err != nil {
				ret = &Return{
					Code:    ErrAPIReadData,
					Message: fmt.Sprintf("read data error: %s", err.Error()),
				}
			} else {
				data = body
			}
		}

		if ret.Code == 0 {
			ret = fn(&Context{w: w, r: r, action: action, data: data})
		}
	}

	if ret.Code == 0 {
		ret.Code = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")

	if retBytes, err := json.Marshal(ret); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":500,"message":"Internal Server Error"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(retBytes)
	}
}

func ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", apiHandler)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return server.ListenAndServe()
}

func ListenAndServeTLS(addr, certFile, keyFile string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", apiHandler)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return server.ListenAndServeTLS(certFile, keyFile)
}

func ListenAndServeTLSWithCert(addr string, certBytes []byte, keyBytes []byte) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", apiHandler)

	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server := &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	return server.ListenAndServeTLS("", "")
}

func RegisterHandler(action string, handler func(ctx *Context) *Return) {
	gMap[action] = handler
}
