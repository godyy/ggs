package actors

import (
	"github.com/godyy/ggs/internal/base/actor"
	"github.com/godyy/ggs/internal/base/actor/model"
)

// GetModule 通过actor获取模块的通用泛型封装.
func GetModule[M model.Module](a actor.ActorWithModule, autoCreate bool) M {
	return actor.GetModule[M](a, autoCreate)
}
