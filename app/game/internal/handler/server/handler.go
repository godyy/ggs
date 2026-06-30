package server

import (
	"github.com/godyy/ggs/app/game/internal/handler"
	actorhandler "github.com/godyy/ggs/internal/infra/actor/handler"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
)

func init() {
	handler.RegisterS2S((*pbs2s.GetServerNameReq)(nil), actorhandler.WrapRPCFunc(handleGetServerName))
}
