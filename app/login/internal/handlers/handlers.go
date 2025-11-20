package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/godyy/ggs/app/login/internal"
	"github.com/godyy/ggs/app/login/internal/handlers/middleware"
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
	root = root.Group("/"+version, middleware.Auth)
	for _, h := range handlers {
		h.setupRoutes(root, version)
	}
}

func init() {
	internal.SetupRoutes = func(root *gin.RouterGroup) {
		version := "v1"
		root = root.Group("/"+version, middleware.Auth)
		for _, h := range handlers {
			h.setupRoutes(root, version)
		}
	}
}
