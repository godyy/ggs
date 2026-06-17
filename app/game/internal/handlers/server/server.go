package server

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/infra/actors"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
	"github.com/godyy/ggskit/infra/actor"
)

func handleGetServerName(ctx *gactor.Context, req *pbs2s.GetServerNameReq) (*pbs2s.GetServerNameResp, error) {
	server := actor.CtxActor[*actors.Server](ctx)
	return &pbs2s.GetServerNameResp{
		ServerName: systems.Server.GetServerName(server),
	}, nil
}
