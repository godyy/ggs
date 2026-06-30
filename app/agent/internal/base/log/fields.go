package log

import (
	"net"

	"github.com/godyy/ggskit/base/protocol"
	"go.uber.org/zap"
)

func FldUid(uid string) zap.Field {
	return zap.String("uid", uid)
}

func FldPlayerId(playerId int64) zap.Field {
	return zap.Int64("playerId", playerId)
}

func FldServerId(serverId int64) zap.Field {
	return zap.Int64("serverId", serverId)
}

func FldRemoteAddr(addr net.Addr) zap.Field {
	return zap.Any("remoteAddr", addr)
}

func FldError(err error) zap.Field {
	return zap.NamedError("err", err)
}

func FldPid(pid protocol.PID) zap.Field {
	return zap.Uint32("pid", uint32(pid))
}

func FldSessionId(sessionId uint32) zap.Field {
	return zap.Uint32("sessionId", sessionId)
}

func FldNodeId(nodeId string) zap.Field {
	return zap.String("nodeId", nodeId)
}
