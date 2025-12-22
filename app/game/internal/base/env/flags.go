package env

import (
	"fmt"

	"github.com/godyy/ggs/internal/base/env"
)

func (e *Env) applyFlags() {
	sid, ok := env.GetFlagValue[int64]("server-id")
	if ok && sid > 0 {
		e.serverId = sid
		e.db = fmt.Sprintf("game_%d", sid)
	} else {
		panic("env: env-server-id is required and must > 0")
	}

	if master, ok := env.GetFlagValue[bool]("master"); ok {
		e.master = master
	}
}

func init() {
	env.AddFlag("server-id", int64(0), "server id")
	env.AddFlag("master", false, "is master node")
}
