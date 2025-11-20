package env

import (
	"github.com/godyy/ggs/internal/env"
)

// accessor 环境变量访问器.
type accessor struct {
	env.Accessor
}

var (
	ac = &accessor{}

	serverId int64 // 服务器ID

	db string // 服务器数据库名称
)

// All 返回所有环境变量.
func All() *accessor {
	return ac
}

// ServerID 服务器ID
func (*accessor) ServerID() int64 {
	return serverId
}

// DB 服务器数据库名称
func (*accessor) DB() string {
	return db
}
