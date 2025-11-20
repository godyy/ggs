package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// handler 路由处理接口.
type handler interface {
	// groupPath 返回路由组路径.
	groupPath() string

	// setupRoutes 配置路由.
	setupRoutes(root *gin.RouterGroup, version string)
}

var (
	handlers = map[string]handler{}
)

func reigsterHandler(h handler) {
	// 检查路由组路径是否已存在
	if _, ok := handlers[h.groupPath()]; ok {
		panic(fmt.Sprintf("handler group path %s already registered", h.groupPath()))
	}
	handlers[h.groupPath()] = h
}

func SetupRoutes(root *gin.RouterGroup, version string) {
	for _, h := range handlers {
		h.setupRoutes(root, version)
	}
}
