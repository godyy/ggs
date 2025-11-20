package player

import (
	"github.com/godyy/ggs/app/game/internal/base/consts"
	"github.com/godyy/ggs/internal/base/actor/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Model 玩家数据模型.
type Model struct {
	*model.ModelDirty `bson:"-"` // 集成DirtyModel

	ID      int64            `bson:"id"`      // 玩家ID
	Version int32            `bson:"version"` // 版本号
	Modules *model.ModuleMgr `bson:"modules"` // 模块数据
}

// New 构造玩家数据.
func New(a model.ActorWithModel, id int64) *Model {
	p := &Model{
		ModelDirty: model.NewModelDirty(a),
		ID:         id,
	}
	p.Modules = model.NewModuleMgr(p)
	return p
}

func (p *Model) GetHashKey() any { return p.ID }

// GetCollection 存储Player的集合名称.
func (p *Model) GetCollection() string { return consts.MgoDBCollPlayers }

// GetFilter 获取Player的查询过滤器.
func (p *Model) GetFilter() any {
	return bson.M{"id": p.ID}
}

// Release 释放玩家数据资源.
func (p *Model) Release() {
	p.ModelDirty.Release()
	p.Modules.Release()
}

// ModuleRegistry 获取模块注册表.
func (p *Model) ModuleRegistry() *model.ModuleRegistry {
	return moduleRegistry
}

// SetModuleDirty 设置模块脏数据.
func (p *Model) SetModuleDirty(m model.Module) {
	p.ModelDirty.SetDirty("modules."+m.ModuleKey(), m)
}

func (p *Model) GetModule(key string, autoCreate bool) model.Module {
	return p.Modules.GetModule(key, autoCreate)
}

// IsInit 是否初始化.
func (p *Model) IsInit() bool {
	return p.Version > 0
}
