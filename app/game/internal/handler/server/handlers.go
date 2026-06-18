package server

import (
	actorhandler "github.com/godyy/ggs/internal/infra/actor/handler"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
	"github.com/godyy/ggskit/infra/actor"
)

var (
	handler = actorhandler.NewS2SHandler()
)

func init() {
	initS2SHandler()
}

func initS2SHandler() {
	registerS2SFunc(pbs2s.PID_PGetServerNameReq, actorhandler.WrapRPCFunc(handleGetServerName))
}

func registerS2SFunc(pid pbs2s.PID, f ...actor.HandlerFunc) {
	handler.RegisterFunc(pid, f...)
}

func Handle(ctx *actor.Context) {
	handler.Handle(ctx)
}
