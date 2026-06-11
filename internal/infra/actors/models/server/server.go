package server

import (
	"fmt"

	"github.com/godyy/ggs/internal/infra/actors/models"
	"github.com/godyy/ggskit/infra/actor"
)

// Model server 数据模型.
type Model struct {
	*actor.ModelDirtyAll   `bson:"-"`
	models.FieldID[string] `bson:",inline"` // 集成通用ID字段

	serverId int64

	Version int32 `json:"version"` // 版本.
}

// New 创建server 数据模型.
func New(a actor.ActorWithModel, serverId int64) *Model {
	m := &Model{
		ModelDirtyAll: actor.NewModelDirtyAll(a),
		serverId:      serverId,
	}
	m.ID = fmt.Sprintf("server_%d", serverId)
	return m
}

// GetHashKey 获取Model的哈希键.
func (m *Model) GetHashKey() any {
	return m.serverId
}

// GetCollection 存储Model的集合名称.
func (m *Model) GetCollection() string {
	return models.CollSingleton
}
