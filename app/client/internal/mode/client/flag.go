package client

import (
	"log"

	"github.com/godyy/ggs/internal/libs/flags"
)

var (
	uid      string // 用户ID
	serverId int64  // 服务器ID
)

func init() {
	flags.String("client-uid", "", "client uid")
	flags.Int64("client-server-id", 0, "client server id")
}

func applyFlags() {
	uid, _ = flags.GetValue[string]("client-uid")
	if uid == "" {
		log.Fatalf("-client-uid is empty")
	}
	serverId, _ = flags.GetValue[int64]("client-server-id")
}
