package env

import (
	"fmt"

	"github.com/godyy/ggs/internal/base/env"
	"github.com/godyy/ggs/internal/libs/flags"
)

func (e *Env) applyFlags() {
	sid, ok := flags.GetValue[int64](env.FlagName("server-id"))
	if ok && sid > 0 {
		e.serverId = sid
		e.db = fmt.Sprintf("game_%d", sid)
	} else {
		panic("env: env-server-id is required and must > 0")
	}
}

func init() {
	flags.Int64(env.FlagName("server-id"), 0, "server id")
}
