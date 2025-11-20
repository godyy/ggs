package env

import (
	"fmt"

	"github.com/godyy/ggs/internal/env"

	"github.com/godyy/ggs/internal/libs/flags"
)

func Init() {
	env.Init()

	sid, ok := flags.GetValue[int64]("env-server-id")
	if ok && sid > 0 {
		serverId = sid
		db = fmt.Sprintf("game_%d", sid)
	} else {
		panic("env: env-server-id is required and must > 0")
	}
}

func init() {
	flags.Int64("env-server-id", 0, "server id")
}
