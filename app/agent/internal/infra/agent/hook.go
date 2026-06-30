package agent

import (
	"reflect"

	pbc2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
	"github.com/godyy/ggs/internal/infra/actor/protocol/registry/c2s"
	"github.com/godyy/ggskit/base/protocol"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type msgHook func(a *Agent, p []byte, msg proto.Message)

var msgHooks map[protocol.PID]msgHook

func init() {
	msgHooks = make(map[protocol.PID]msgHook, 4)
	registerMsgHook((*pbc2s.LoginReq)(nil), handleLoginReq)
	registerMsgHook((*pbc2s.LoginCharacterResp)(nil), handleLoginGameResp)
}

func registerMsgHook(msg proto.Message, hook msgHook) {
	pid, ok := c2s.Registry.GetPid(msg)
	if !ok {
		panic(pkgerrors.Errorf("hook message not registered: %s", reflect.TypeOf(msg)))
	}
	msgHooks[pid] = hook
}

func isMsgHooked(pid protocol.PID) bool {
	_, ok := msgHooks[pid]
	return ok
}

func getMsgHook(pid protocol.PID) msgHook {
	return msgHooks[pid]
}
