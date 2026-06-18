package player

import (
	"time"

	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	"github.com/godyy/ggs/internal/protocol/pb/s2s"
	pkgerrors "github.com/pkg/errors"
)

// handleLoginCharacter 处理登录角色请求.
func handleLoginCharacter(ctx *actor.Context, req *pbc2s.LoginCharacterReq) (*pbc2s.LoginCharacterResp, error) {
	player := actor.CtxActor[*actors.Player](ctx)

	if err := systems.Player.InitPlayer(player); err != nil {
		return nil, err
	}

	player.SetLogin()

	getServerNameResp, err := player.Sugared().RPCWithTimeout(actor.ActorUID{Category: actor.CategoryServer.ActorCategory(), ID: 1},
		&s2s.GetServerNameReq{}, 5*time.Second)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "get server name")
	}
	logger.Get().Info("get server name success, server name: %s", getServerNameResp.(*s2s.GetServerNameResp).ServerName)

	return &pbc2s.LoginCharacterResp{}, nil
}

// handleHearbeat 处理心跳.
func handleHeartbeat(ctx *actor.Context, req *pbc2s.HeartbeatReq) (*pbc2s.HeartbeatResp, error) {
	player := actor.CtxActor[*actors.Player](ctx)
	player.Heartbeat()
	return &pbc2s.HeartbeatResp{}, nil
}
