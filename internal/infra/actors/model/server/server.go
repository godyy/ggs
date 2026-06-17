package server

import (
	"fmt"

	"github.com/godyy/ggs/internal/infra/actors/model"
	"github.com/godyy/ggskit/infra/actor"
)

// Model server 数据模型.
type Model struct {
	model.Dirty           `bson:"-"`       // 集成脏标记位
	model.FieldID[string] `bson:",inline"` // 集成通用ID字段

	serverId int64

	Version    int32  `bson:"version"`    // 版本.
	ServerName string `bson:"serverName"` // 服务器名.
}

// New 创建server 数据模型.
func New(a actor.ActorWithModel, serverId int64) *Model {
	m := &Model{
		serverId:   serverId,
		ServerName: fmt.Sprintf("server%d", serverId),
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
	return model.CollSingleton
}
