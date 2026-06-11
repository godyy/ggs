package env

import baseenv "github.com/godyy/ggskit/base/env"

// Env 环境变量管理器.
type Env struct {
	baseenv.Env

	serverId int64 // 服务器 ID
}

// NewEnv 创建环境变量管理器.
func NewEnv() *Env {
	return &Env{
		Env: baseenv.Get(),
	}
}

// Init 初始化环境变量.
func (e *Env) Init() {
	e.applyFlags()
}

// ServerID 返回服务器 ID.
func (e *Env) ServerID() int64 {
	return e.serverId
}
