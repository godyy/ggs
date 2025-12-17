package probe

import (
	"github.com/gin-gonic/gin"
)

// RegisterGin 注册 gin 路由，默认路径为 /healthz 与 /readyz。
func RegisterGin(engine *gin.Engine, healthPath, readyPath string) {
	if healthPath == "" {
		healthPath = "/healthz"
	}
	if readyPath == "" {
		readyPath = "/readyz"
	}
	engine.GET(healthPath, func(c *gin.Context) { HealthzHandler()(c.Writer, c.Request) })
	engine.GET(readyPath, func(c *gin.Context) { ReadyzHandler()(c.Writer, c.Request) })
}
