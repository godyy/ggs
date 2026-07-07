package env

import (
	"github.com/godyy/ggskit/base/env"
)

// Env 环境变量管理器.
type Env struct {
	env.Env

	serverId int64 // 服务器ID
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
