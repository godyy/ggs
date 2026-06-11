package player

import (
	"github.com/godyy/ggs/internal/infra/actors/models"
	"github.com/godyy/ggskit/infra/actor"
)

const DBColl = "players"

// Model 玩家数据模型.
type Model struct {
	*actor.ModelDirty     `bson:"-"`       // 集成DirtyModel
	models.FieldID[int64] `bson:",inline"` // 集成通用ID字段, 即玩家ID, ActorUID.ID

	Version int32            `bson:"version"` // 版本号
	Modules *actor.ModuleMgr `bson:"modules"` // 模块数据
}

// New 构造玩家数据.
func New(a actor.ActorWithModule, id int64) *Model {
	p := &Model{
		ModelDirty: actor.NewModelDirty(a),
	}
	p.ID = id
	p.Modules = actor.NewModuleMgr(p)
	return p
}

func (p *Model) GetHashKey() any { return p.ID }

// GetCollection 存储Player的集合名称.
func (p *Model) GetCollection() string { return DBColl }

// Release 释放玩家数据资源.
func (p *Model) Release() {
	p.ModelDirty.Release()
	p.Modules.Release()
}

// ModuleRegistry 获取模块注册表.
func (p *Model) ModuleRegistry() *actor.ModuleRegistry {
	return moduleRegistry
}

// SetModuleDirty 设置模块脏数据.
func (p *Model) SetModuleDirty(key string) {
	if m := p.Modules.GetModule(key, false); m != nil {
		p.ModelDirty.SetDirty("modules."+m.ModuleKey(), m)
	}
}

func (p *Model) GetModule(key string, autoCreate bool) actor.Module {
	return p.Modules.GetModule(key, autoCreate)
}

// IsInit 是否初始化.
func (p *Model) IsInit() bool {
	return p.Version > 0
}
