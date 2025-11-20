package actors

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/actors/model"
)

// ActorCouldPersist 可持久化Actor接口.
type ActorCouldPersist interface {
	gactor.ActorBehavior

	// AsyncSaveAll 异步保存所有数据.
	AsyncSaveAll()
}

// persistor 持久化辅助结构
type persistor struct {
	saveTimerId gactor.TimerId // save 定时器ID
}

func (p *persistor) SaveTimerId() gactor.TimerId {
	return p.saveTimerId
}

func (p *persistor) SetSaveTimerId(timerId gactor.TimerId) {
	p.saveTimerId = timerId
}

// ActorWithModule 包含模块的Actor接口.
type ActorWithModule interface {
	gactor.Actor
	model.ModuleGetter
}

// GetModule 通过actor获取模块的通用泛型封装.
func GetModule[M model.Module](actor ActorWithModule, autoCreate bool) M {
	return model.GetModule[M](actor, autoCreate)
}
