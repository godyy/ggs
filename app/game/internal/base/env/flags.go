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
	} else {
		panic("env: env-server-id is required and must > 0")
	}

	gdconfDB, ok := env.GetFlagValue[string]("gdconf-db")
	if ok && gdconfDB != "" {
		e.gdconfDB = gdconfDB
	}
}

func init() {
	env.AddFlag("server-id", int64(0), "server id")
	env.AddFlag("gdconf-db", "gdconf_dev", "gdconf db name")
}
