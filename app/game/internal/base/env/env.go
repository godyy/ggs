package env

import (
	"errors"

	"github.com/godyy/ggskit/base/env"
)

// Role 描述 game 进程的运行职责.
type Role string

const (
	RoleFixed  Role = "fixed"
	RoleWorker Role = "worker"
)

// WorkerState 描述 worker 启动后的玩家接入状态.
type WorkerState string

const (
	WorkerStateStandby WorkerState = "standby"
	WorkerStateServing WorkerState = "serving"
)

// Env 环境变量管理器.
type Env struct {
	env.Env

	serverId    int64       // 服务器ID
	db          string      // 服务器数据库名称
	gdconfDB    string      // gdconf数据库名称
	role        Role        // 进程角色
	workerID    string      // worker 节点名称
	workerState WorkerState // worker 初始接入状态
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

// DB 服务器数据库名称
func (e *Env) DB() string {
	return e.db
}

// GDconfDB gdconf数据库名称
func (e *Env) GDconfDB() string {
	return e.gdconfDB
}

// Role 返回进程运行角色.
func (e *Env) Role() Role {
	return e.role
}

// IsWorker 是否为 Player worker.
func (e *Env) IsWorker() bool {
	return e.role == RoleWorker
}

// WorkerID 返回 worker 节点名称.
func (e *Env) WorkerID() string {
	return e.workerID
}

// WorkerState 返回 worker 初始接入状态.
func (e *Env) WorkerState() WorkerState {
	return e.workerState
}

func parseRole(value string) (Role, error) {
	role := Role(value)
	switch role {
	case RoleFixed, RoleWorker:
		return role, nil
	default:
		return "", errors.New("env-role must be fixed or worker")
	}
}

// parseWorkerState 解析 worker 允许的初始接入状态。
func parseWorkerState(value string) (WorkerState, error) {
	state := WorkerState(value)
	switch state {
	case WorkerStateStandby, WorkerStateServing:
		return state, nil
	default:
		return "", errors.New("env-state must be standby or serving")
	}
}
