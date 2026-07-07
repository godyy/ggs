package env

import (
	"github.com/godyy/ggskit/base/env"
)

func (e *Env) applyFlags() {
	sid, ok := env.GetFlagValue[int64]("server-id")
	if ok && sid > 0 {
		e.serverId = sid
	} else {
		panic("env: env-server-id is required and must > 0")
	}
}

func init() {
	env.AddFlag("server-id", int64(0), "server id")
}
