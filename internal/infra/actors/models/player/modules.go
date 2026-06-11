package player

import (
	"github.com/godyy/ggskit/infra/actor"
)

// moduleBase 模块基础别名
// 方便嵌套时不用手动添加bson屏蔽标签.
type moduleBase[M actor.Module] = actor.ModuleBase[M]

// moduleRegistry 模块注册表
var moduleRegistry = actor.NewModuleRegistry()

func init() {
	// 注册模块
	actor.RegisterModule[*BaseInfo](moduleRegistry)
}
