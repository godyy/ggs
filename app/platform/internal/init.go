package internal

import (
	"github.com/gin-gonic/gin"
	_ "github.com/godyy/ggs/internal/infra/actor/actors"
)

var SetupRoutes func(root *gin.RouterGroup)
