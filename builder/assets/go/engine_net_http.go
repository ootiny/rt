package _rt_package_name_

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GoContext struct {
	action string
	r      *http.Request
	ContextBase
}

func (p *GoContext) Action() string {
	return p.action
}

func (p *GoContext) Method() string {
	return p.r.Method
}

func (p *GoContext) Cookie(name string) (*http.Cookie, error) {
	return p.r.Cookie(name)
}

func (p *GoContext) Header(name string) string {
	return p.r.Header.Get(name)
}

type GoResponse struct {
	w http.ResponseWriter
}

func (p *GoResponse) WriteJson(data []byte) (int, error) {
	return p.w.Write(data)
}

func NewHttpServer(addr string, certFile string, keyFile string) *Server {
	mux := http.NewServeMux()

	// API handler function with CORS support
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

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
				ret = fn(&GoContext{action: action, r: r}, &GoResponse{w: w}, data)
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

	mux.HandleFunc("/api", apiHandler)

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
