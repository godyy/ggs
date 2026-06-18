package server

import (
	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
)

func handleGetServerName(ctx *actor.Context, req *pbs2s.GetServerNameReq) (*pbs2s.GetServerNameResp, error) {
	server := actor.CtxActor[*actors.Server](ctx)
	return &pbs2s.GetServerNameResp{
		ServerName: systems.Server.GetServerName(server),
	}, nil
}
