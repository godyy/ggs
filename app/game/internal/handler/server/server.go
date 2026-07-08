package server

import (
	"github.com/godyy/ggs/app/game/internal/app"
	"github.com/godyy/ggs/app/game/internal/handler"
	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/gdconf"
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

func handleReloadGDConf(ctx *actor.Context, req *pbs2s.ReloadGDConfReq) (*pbs2s.ReloadGDConfResp, error) {
	resp := &pbs2s.ReloadGDConfResp{
		Success: true,
	}
	db := app.MongoClient().Database(app.Env().GDconfDB())
	if req.All {
		if err := gdconf.Load(db); err != nil {
			resp.Success = false
			resp.Err = err.Error()
			handler.Logger().Errorf("reload gdconf, %v", err)
		} else {
			handler.Logger().Info("reload gdconf")
		}
	} else if len(req.Tables) > 0 {
		if err := gdconf.LoadTable(db, req.Tables...); err != nil {
			resp.Success = false
			resp.Err = err.Error()
			handler.Logger().Errorf("reload gdconf tables %v, %v", req.Tables, err)
		} else {
			handler.Logger().Info("reload gdconf tables %v", req.Tables)
		}
	}
	return resp, nil
}
