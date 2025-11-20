package player

import (
	"github.com/godyy/ggs/app/game/internal/actors/model"
)

// moduleBase 模块基础别名
// 方便嵌套时不用手动添加bson屏蔽标签.
type moduleBase[M model.Module] = model.ModuleBase[M]

// moduleRegistry 模块注册表
var moduleRegistry = model.NewModuleRegistry()

func init() {
	// 注册模块
	model.RegisterModule[*BaseInfo](moduleRegistry)
}
