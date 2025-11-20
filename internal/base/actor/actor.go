package actor

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/actor/model"
)

type Actor = gactor.Actor

type CActor = gactor.CActor

// ActorWithModel 包含模型的Actor接口.
type ActorWithModel interface {
	Actor

	// GetModel 获取模型实例.
	GetModel() model.Model
}

// ActorWithModule 包含数据模块的Actor接口.
type ActorWithModule interface {
	ActorWithModel
	model.ModuleGetter
}

// GetModule 通过actor获取模块的通用泛型封装.
func GetModule[M model.Module](actor ActorWithModule, autoCreate bool) M {
	return model.GetModule[M](actor, autoCreate)
}
