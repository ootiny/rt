package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Context struct {
	action string
	r      *http.Request
	ContextBase
}

func (p *Context) Action() string {
	return p.action
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

type Response struct {
	w http.ResponseWriter
}

func (p *Response) WriteJson(data []byte) (int, error) {
	return p.w.Write(data)
}

func NewHttpServer(addr string, certFile string, keyFile string) IServer {
	mux := http.NewServeMux()

	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("a")
		ret := (*Return)(nil)
		fn, ok := gHttpActionMap[action]

		if !ok {
			ret = &Return{
				Code:    ErrActionNotFound,
				Message: fmt.Sprintf("action %s not found", action),
			}
		} else {
			data := []byte(r.URL.Query().Get("d"))

			if len(data) == 0 {
				if body, err := io.ReadAll(r.Body); err != nil {
					ret = &Return{
						Code:    ErrReadData,
						Message: fmt.Sprintf("read data error: %s", err.Error()),
					}
				} else {
					data = body
				}
			}

			if ret.Code == 0 {
				ret = fn(&Context{action: action, r: r}, &Response{w: w}, data)
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
	})

	return &Server{
		addr:     addr,
		certFile: certFile,
		keyFile:  keyFile,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

type Server struct {
	addr     string
	certFile string
	keyFile  string
	server   *http.Server
}

func (p *Server) Run() error {
	if p.certFile != "" || p.keyFile != "" {
		return p.server.ListenAndServeTLS(p.certFile, p.keyFile)
	} else {
		return p.server.ListenAndServe()
	}
}
