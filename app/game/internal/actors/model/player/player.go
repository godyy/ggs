package player

import (
	"github.com/godyy/ggs/app/game/internal/actors/model"
	"github.com/godyy/ggs/app/game/internal/consts"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Player 玩家数据.
type Player struct {
	*model.DirtyModel `bson:"-"` // 集成DirtyModel

	ID      int64            `bson:"id"`      // 玩家ID
	Version int32            `bson:"version"` // 版本号
	Modules *model.ModuleMgr `bson:"modules"` // 模块数据
}

// NewPlayer 构造玩家数据.
func NewPlayer(a model.Actor, id int64) *Player {
	p := &Player{
		DirtyModel: model.NewDirtyModel(a),
		ID:         id,
	}
	p.Modules = model.NewModuleMgr(p)
	return p
}

func (p *Player) GetHashKey() any { return p.ID }

// GetCollection 存储Player的集合名称.
func (p *Player) GetCollection() string { return consts.MgoDBCollPlayers }

// GetFilter 获取Player的查询过滤器.
func (p *Player) GetFilter() any {
	return bson.M{"id": p.ID}
}

// Release 释放玩家数据资源.
func (p *Player) Release() {
	p.DirtyModel.Release()
	p.Modules.Release()
}

// ModuleRegistry 获取模块注册表.
func (p *Player) ModuleRegistry() *model.ModuleRegistry {
	return moduleRegistry
}

// SetModuleDirty 设置模块脏数据.
func (p *Player) SetModuleDirty(m model.Module) {
	p.DirtyModel.SetDirty("modules."+m.Key(), m)
}

func (p *Player) GetModule(key string, autoCreate bool) model.Module {
	return p.Modules.GetModule(key, autoCreate)
}

// IsInit 是否初始化.
func (p *Player) IsInit() bool {
	return p.Version > 0
}
