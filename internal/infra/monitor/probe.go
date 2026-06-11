package monitor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/godyy/ggskit/infra/monitor/probe"
)

func InitProbeHttp(mux *http.ServeMux, basePath string, opts ...probe.Option) {
	router := NewHttpRouter(mux, basePath)
	probe.Init(router, opts...)
}

func InitProbeGin(router *gin.RouterGroup, opts ...probe.Option) {
	GinRouter := NewGinRouter(router)
	probe.Init(GinRouter, opts...)
}
