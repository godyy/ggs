package env

import baseenv "github.com/godyy/ggskit/base/env"

func (e *Env) applyFlags() {
	serverId, ok := baseenv.GetFlagValue[int64]("server-id")
	if !ok || serverId <= 0 {
		panic("env: env-server-id is required and must > 0")
	}

	e.serverId = serverId
}

func init() {
	baseenv.AddFlag("server-id", int64(0), "server id")
}
