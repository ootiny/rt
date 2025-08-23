package _rt_package_name_

import (
	"io"
	"net/http"
)

type GoRequest struct {
	r *http.Request
}

func (p *GoRequest) Action() string {
	return p.r.URL.Query().Get("a")
}

func (p *GoRequest) Data() []byte {
	if p.r.Method == http.MethodGet {
		return []byte(p.r.URL.Query().Get("d"))
	} else {
		body, err := io.ReadAll(p.r.Body)
		if err != nil {
			return nil
		}
		return body
	}
}

func (p *GoRequest) Cookie(name string) (*http.Cookie, error) {
	return p.r.Cookie(name)
}

func (p *GoRequest) Header(name string) string {
	return p.r.Header.Get(name)
}

type GoResponse struct {
	w http.ResponseWriter
}

func (p *GoResponse) WriteJson(data []byte) (int, error) {
	return p.w.Write(data)
}

func (p *GoResponse) SetHeader(name string, value string) {
	p.w.Header().Set(name, value)
}

func (p *GoResponse) WriteHeader(code int) {
	p.w.WriteHeader(code)
}

func NewHttpServer(addr string, certFile string, keyFile string, cors bool) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		apiHandler(cors, &GoResponse{w: w}, &GoRequest{r: r})
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
