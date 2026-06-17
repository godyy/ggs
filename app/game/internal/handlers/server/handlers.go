package server

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/handlers"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
)

var (
	handler = handlers.NewHandler()
)

func init() {
	initS2SHandler()
}

func initS2SHandler() {
	registerS2SFunc(pbs2s.PID_PGetServerNameReq, handlers.WrapS2SRPCFunc(handleGetServerName))
}

func registerS2SFunc(pid pbs2s.PID, f ...gactor.HandlerFunc) {
	handlers.RegisterS2SFunc(handler, pid, f...)
}

func Handle(ctx *gactor.Context) {
	handler.Handle(ctx)
}
