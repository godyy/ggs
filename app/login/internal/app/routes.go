package app

import (
	"github.com/gin-gonic/gin"
	"github.com/godyy/ggs/app/login/internal/handlers"
	"github.com/godyy/ggs/app/login/internal/handlers/middleware"
)

// 配置路由
func setupRoutes(engine *gin.Engine) {
	api_v1 := engine.Group("/api/v1", middleware.Auth)
	handlers.SetupRoutes(api_v1, "v1")
}
