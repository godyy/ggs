package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/handlers"
	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
	pbs2s "github.com/godyy/ggs/internal/proto/pb/s2s"
)

var (
	c2sHandler = handlers.NewHandler()
	s2sHandler = handlers.NewHandler()
)

func init() {
	initC2SHandler()
	initS2SHandler()
}

func initC2SHandler() {
	registerC2SFunc(pbc2s.PID_PLoginCharacterReq, false, handlers.WrapC2SFunc(handleLoginCharacter))
	registerC2SFunc(pbc2s.PID_PHeartbeatReq, true, handlers.WrapC2SFunc(handleHeartbeat))
	registerC2SFunc(pbc2s.PID_PModifyNameReq, true, handlers.WrapC2SFunc(handleModifyName))
}

func initS2SHandler() {
	registerS2SFunc(pbs2s.PID_PActorSaveResult, handlers.WrapS2SCastFunc(handlers.OnActorSaveResult))
}

func registerC2SFunc(pid pbc2s.PID, checkLogin bool, f ...gactor.HandlerFunc) {
	if checkLogin {
		handlers.RegisterC2SFunc(c2sHandler, pid, append([]gactor.HandlerFunc{mdCheckLogin}, f...)...)
	} else {
		handlers.RegisterC2SFunc(c2sHandler, pid, f...)
	}
}

func registerS2SFunc(pid pbs2s.PID, f ...gactor.HandlerFunc) {
	handlers.RegisterS2SFunc(s2sHandler, pid, f...)
}

func Handle(ctx *gactor.Context) {
	if ctx.RequestType() == gactor.RequestTypeReq {
		c2sHandler.Handle(ctx)
	} else {
		s2sHandler.Handle(ctx)
	}
}
