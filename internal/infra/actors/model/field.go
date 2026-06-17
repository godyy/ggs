package model

import (
	"github.com/godyy/ggskit/infra/actor"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ID 通用ID类型约束.
type ID interface {
	int8 | int16 | int32 | int64 | int | uint8 | uint16 | uint32 | uint64 | uint | string
}

// FieldID 通用泛型字段ID.
type FieldID[T ID] struct {
	ID T `bson:"_id"`
}

func (f *FieldID[T]) GetFilter() any {
	return bson.M{"_id": f.ID}
}

// FieldModules 数据模块集合字段.
type FieldModules struct {
	Modules *actor.Modules `bson:"modules"`
}

// NewFieldModules 创建数据模块集合字段.
func NewFieldModules(mr *actor.ModuleRegistry) *FieldModules {
	return &FieldModules{Modules: actor.NewModules(mr)}
}

// GetBsonTag 获取数据模块集合字段的BSON标签.
func (f *FieldModules) GetBsonTag() string {
	return "modules"
}

// GetModule 获取数据模块.
func (f *FieldModules) GetModule(mk actor.ModuleKey, autoCreate bool) actor.Module {
	return f.Modules.GetModule(mk, autoCreate)
}
