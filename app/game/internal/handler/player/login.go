package player

import (
	"time"

	"github.com/godyy/ggs/app/game/internal/app"
	"github.com/godyy/ggs/app/game/internal/base/event"
	"github.com/godyy/ggs/app/game/internal/global"
	"github.com/godyy/ggs/app/game/internal/handler"
	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	pbc2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
	"github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
	pkgerrors "github.com/pkg/errors"
)

// handleLoginCharacter 处理登录角色请求.
func handleLoginCharacter(ctx *actor.Context, req *pbc2s.LoginCharacterReq) (*pbc2s.LoginCharacterResp, error) {
	player := actor.CtxActor[*actors.Player](ctx)

	if err := systems.Player.InitPlayer(player); err != nil {
		return nil, err
	}

	player.SetLogin()
	if err := global.NewEventWrapperNoParams[*actors.Player]().DispatchKind(event.KindPlayerOnline, player); err != nil {
		handler.Logger().Errorf("dispatch player online event, err:%v", err)
		return nil, err
	}

	getServerNameResp, err := player.RPCWithTimeout(actor.ActorUID{Category: actor.CategoryServer.ActorCategory(), ID: app.Env().ServerID()},
		&s2s.GetServerNameReq{}, 5*time.Second)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "get server name")
	}
	handler.Logger().Info("get server name success, server name: %s", getServerNameResp.(*s2s.GetServerNameResp).ServerName)

	return &pbc2s.LoginCharacterResp{}, nil
}

// handleHearbeat 处理心跳.
func handleHeartbeat(ctx *actor.Context, req *pbc2s.HeartbeatReq) (*pbc2s.HeartbeatResp, error) {
	player := actor.CtxActor[*actors.Player](ctx)
	player.Heartbeat()
	return &pbc2s.HeartbeatResp{}, nil
}
