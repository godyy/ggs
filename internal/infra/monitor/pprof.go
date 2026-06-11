package monitor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/godyy/ggskit/infra/monitor/pprof"
)

func RegisterPProfHttp(mux *http.ServeMux, basePath string) {
	router := NewHttpRouter(mux, basePath)
	pprof.RegisterHandler(router, "")
}

func RegisterPProfGin(router *gin.RouterGroup) {
	ginRouter := NewGinRouter(router)
	pprof.RegisterHandler(ginRouter, "")
}
