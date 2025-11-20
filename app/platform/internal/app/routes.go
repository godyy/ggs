package app

import (
	"github.com/gin-gonic/gin"
	"github.com/godyy/ggs/app/platform/internal/handlers"
)

// 配置路由
func setupRoutes(engine *gin.Engine) {
	api_v1 := engine.Group("/api/v1")
	handlers.SetupRoutes(api_v1, "v1")
}
