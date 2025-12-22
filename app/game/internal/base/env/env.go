package env

import (
	"github.com/godyy/ggs/internal/base/env"
)

// Env 环境变量管理器.
type Env struct {
	env.Env

	serverId int64  // 服务器ID
	master   bool   // 是否为主节点
	db       string // 服务器数据库名称
}

// NewEnv 创建环境变量管理器.
func NewEnv() *Env {
	return &Env{
		Env: env.Get(),
	}
}

// Init 初始化环境变量.
func (e *Env) Init() {
	e.applyFlags()
}

// ServerID 服务器ID
func (e *Env) ServerID() int64 {
	return e.serverId
}

// Master 是否为主节点
func (e *Env) Master() bool {
	return e.master
}

// DB 服务器数据库名称
func (e *Env) DB() string {
	return e.db
}
