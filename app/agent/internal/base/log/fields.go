package log

import (
	"net"

	"go.uber.org/zap"
)

func FldUid(uid string) zap.Field {
	return zap.String("uid", uid)
}

func FldPlayerId(playerId int64) zap.Field {
	return zap.Int64("playerId", playerId)
}

func FldRemoteAddr(addr net.Addr) zap.Field {
	return zap.Any("remoteAddr", addr)
}

func FldError(err error) zap.Field {
	return zap.NamedError("err", err)
}

func FldPid(pid uint16) zap.Field {
	return zap.Uint16("pid", pid)
}

func FldSessionId(sessionId uint32) zap.Field {
	return zap.Uint32("sessionId", sessionId)
}
