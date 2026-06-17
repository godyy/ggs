package player

import (
	"github.com/godyy/ggs/internal/infra/actors/model"
)

const DBColl = "players"

// Model 玩家数据模型.
type Model struct {
	model.ModelWithModules[int64] `bson:",inline"` // 集成模块字段的通用模型

	Version int32 `bson:"version"` // 版本号
}

// New 构造玩家数据.
func New(id int64) *Model {
	p := &Model{
		ModelWithModules: model.NewModelWithModules(id, moduleRegistry),
	}
	return p
}

// GetCollection 存储Player的集合名称.
func (p *Model) GetCollection() string { return DBColl }

// IsInit 是否初始化.
func (p *Model) IsInit() bool {
	return p.Version > 0
}
