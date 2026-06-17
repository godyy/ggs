package model

import (
	"github.com/godyy/ggskit/infra/actor"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ID 通用ID类型约束.
type ID interface {
	int8 | int16 | int32 | int64 | int | uint8 | uint16 | uint32 | uint64 | uint | string
}

// ModelWithID 具备ID字段的模型泛型封装.
type ModelWithID[TID ID] struct {
	ID TID `bson:"_id"` // ID.
}

func NewModuleWithID[TID ID](id TID) ModelWithID[TID] {
	return ModelWithID[TID]{ID: id}
}

func (m *ModelWithID[T]) GetFilter() any {
	return bson.M{"_id": m.ID}
}

func (m *ModelWithID[T]) GetHashKey() any {
	return m.ID
}

func (m *ModelWithID[TID]) Release() {}

// ModelWithModules 具备模块字段的模型泛型封装.
type ModelWithModules[TID ID] struct {
	Dirties          `bson:"-"`       // 集成脏数据搜集器
	ModelWithID[TID] `bson:",inline"` // 集成具备ID字段的通用模型
	Modules          *actor.Modules   `bson:"modules"` // 模块数据
}

func NewModelWithModules[TID ID](id TID, mr *actor.ModuleRegistry) ModelWithModules[TID] {
	return ModelWithModules[TID]{
		Dirties:     NewDirties(),
		ModelWithID: NewModuleWithID(id),
		Modules:     actor.NewModules(mr),
	}
}

func (m *ModelWithModules[TID]) GetModulesBsonTag() string {
	return "modules"
}

func (m *ModelWithModules[TID]) GetModule(mk actor.ModuleKey, autoCreate bool) actor.Module {
	return m.Modules.GetModule(mk, autoCreate)
}

// SetDirtyModule 设置脏模块.
func (p *ModelWithModules[TID]) SetDirtyModule(mk actor.ModuleKey) {
	if m := p.Modules.GetModule(mk, false); m != nil {
		p.Dirties.SetDirty(p.GetModulesBsonTag()+"."+mk.ModuleKey(), m)
	}
}

func (p *ModelWithModules[TID]) Release() {
	p.Dirties.Release()
	p.Modules.Release()
}
