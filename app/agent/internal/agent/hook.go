package agent

import (
	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
	"google.golang.org/protobuf/proto"
)

type msgHook func(a *Agent, p []byte, msg proto.Message)

var msgHooks = map[uint16]msgHook{
	uint16(pbc2s.PID_PLoginReq):           handleLoginReq,
	uint16(pbc2s.PID_PLoginCharacterResp): handleLoginGameResp,
}

func isMsgHooked(pid uint16) bool {
	_, ok := msgHooks[pid]
	return ok
}

func getMsgHook(pid uint16) msgHook {
	return msgHooks[pid]
}
