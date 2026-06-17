package player

import (
	"github.com/godyy/ggs/internal/infra/actors/model"
	"github.com/godyy/ggskit/infra/actor"
)

const DBColl = "players"

// Model 玩家数据模型.
type Model struct {
	model.Dirties        `bson:"-"`       // 集成脏数据搜集器.
	model.FieldID[int64] `bson:",inline"` // 集成通用ID字段, 即玩家ID, ActorUID.ID

	Version              int32            `bson:"version"` // 版本号
	*model.FieldModules `bson:",inline"` // 模块数据
}

// New 构造玩家数据.
func New(id int64) *Model {
	p := &Model{
		Dirties:      model.NewDirties(),
		FieldModules: model.NewFieldModules(moduleRegistry),
	}
	p.ID = id
	return p
}

func (p *Model) GetHashKey() any { return p.ID }

// GetCollection 存储Player的集合名称.
func (p *Model) GetCollection() string { return DBColl }

// Release 释放玩家数据资源.
func (p *Model) Release() {
	p.Dirties.Release()
	p.Modules.Release()
}

// SetDirtyModule 设置脏模块.
func (p *Model) SetDirtyModule(mk actor.ModuleKey) {
	if m := p.Modules.GetModule(mk, false); m != nil {
		p.Dirties.SetDirty(p.FieldModules.GetBsonTag()+"."+mk.ModuleKey(), m)
	}
}

// IsInit 是否初始化.
func (p *Model) IsInit() bool {
	return p.Version > 0
}
