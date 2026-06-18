package player

import (
	"github.com/godyy/ggs/app/game/internal/handler"
	actorhandler "github.com/godyy/ggs/internal/infra/actor/handler"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
)

func init() {
	initC2SHandler()
	initS2SHandler()
}

func initC2SHandler() {
	registerC2SFunc(pbc2s.PID_PLoginCharacterReq, false, actorhandler.WrapReqFunc(handleLoginCharacter))
	registerC2SFunc(pbc2s.PID_PHeartbeatReq, true, actorhandler.WrapReqFunc(handleHeartbeat))
	registerC2SFunc(pbc2s.PID_PModifyNameReq, true, actorhandler.WrapReqFunc(handleModifyName))
	registerC2SFunc(pbc2s.PID_PUseItemReq, true, actorhandler.WrapReqFunc(handleUseItem))
}

func initS2SHandler() {
}

func registerC2SFunc(pid pbc2s.PID, checkLogin bool, f ...actorhandler.HandlerFunc) {
	if checkLogin {
		handler.RegisterC2S(pid, append([]actorhandler.HandlerFunc{mdCheckLogin}, f...)...)
	} else {
		handler.RegisterC2S(pid, f...)
	}
}

func registerS2SFunc(pid pbs2s.PID, f ...actorhandler.HandlerFunc) {
	handler.RegisterS2S(pid, f...)
}
