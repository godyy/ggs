package env

import (
	"fmt"

	"github.com/godyy/ggskit/base/env"
)

func (e *Env) applyFlags() {
	sid, ok := env.GetFlagValue[int64]("server-id")
	if ok && sid > 0 {
		e.serverId = sid
		e.db = fmt.Sprintf("game_%d", sid)
	}

	roleValue, ok := env.GetFlagValue[string]("role")
	if !ok {
		panic("env: env-role is required")
	}
	role, err := parseRole(roleValue)
	if err != nil {
		panic(err)
	}
	e.role = role

	// 固定 game 节点承载区服级 Actor，必须绑定一个区服数据库。
	if !e.IsWorker() && e.serverId <= 0 {
		panic("env: env-server-id is required and must > 0")
	}

	gdconfDB, ok := env.GetFlagValue[string]("gdconf-db")
	if ok && gdconfDB != "" {
		e.gdconfDB = gdconfDB
	}

	workerID, _ := env.GetFlagValue[string]("worker-id")
	if e.IsWorker() {
		if workerID == "" {
			panic("env: env-worker-id is required for worker")
		}
		e.workerID = workerID
		workerState, _ := env.GetFlagValue[string]("state")
		state, err := parseWorkerState(workerState)
		if err != nil {
			panic(err)
		}
		e.workerState = state
	} else if workerID != "" {
		panic("env: env-worker-id is only allowed for worker")
	}
}

func init() {
	env.AddFlag("server-id", int64(0), "server id")
	env.AddFlag("gdconf-db", "gdconf_dev", "gdconf db name")
	env.AddFlag("role", "", "game role: fixed or worker")
	env.AddFlag("worker-id", "", "worker id")
	env.AddFlag("state", "standby", "worker state: standby or serving")
}
