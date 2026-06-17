package player

import (
	"github.com/godyy/ggskit/infra/actor"
)

// moduleRegistry 模块注册表
var moduleRegistry = actor.NewModuleRegistry()

func init() {
	// 注册模块
	actor.RegisterModule[*BaseInfo](moduleRegistry)
	actor.RegisterModule[*Items](moduleRegistry)
}
