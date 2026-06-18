package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/handler"
	"github.com/godyy/ggs/internal/infra/actor"
	actorhandler "github.com/godyy/ggs/internal/infra/actor/handler"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
)

var (
	c2sHandler = actorhandler.NewC2SHandler()
	s2sHandler = actorhandler.NewS2SHandler()
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
	registerS2SFunc(pbs2s.PID_PActorSaveResult, actorhandler.WrapCastFunc(handler.OnActorSaveResult))
}

func registerC2SFunc(pid pbc2s.PID, checkLogin bool, f ...actorhandler.HandlerFunc) {
	if checkLogin {
		c2sHandler.RegisterFunc(pid, append([]actorhandler.HandlerFunc{mdCheckLogin}, f...)...)
	} else {
		c2sHandler.RegisterFunc(pid, f...)
	}
}

func registerS2SFunc(pid pbs2s.PID, f ...actorhandler.HandlerFunc) {
	s2sHandler.RegisterFunc(pid, f...)
}

func Handle(ctx *actor.Context) {
	if ctx.RequestType() == gactor.RequestTypeReq {
		c2sHandler.Handle(ctx)
	} else {
		s2sHandler.Handle(ctx)
	}
}
