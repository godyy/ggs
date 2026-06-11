package monitor

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HttpRouter struct {
	mux      *http.ServeMux
	basePath string
}

func NewHttpRouter(mux *http.ServeMux, basePath string) *HttpRouter {
	return &HttpRouter{mux: mux, basePath: basePath}
}

func (r *HttpRouter) Handle(method, path string, handler http.Handler) {
	r.mux.Handle(method+" "+r.basePath+path, handler)
}

type GinRouter struct {
	router *gin.RouterGroup
}

func NewGinRouter(router *gin.RouterGroup) *GinRouter {
	return &GinRouter{router: router}
}

func (r *GinRouter) Handle(method, path string, handler http.Handler) {
	r.router.Handle(method, path, gin.WrapH(handler))
}
