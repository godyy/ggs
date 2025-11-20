package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/actors"
	"github.com/godyy/ggs/app/game/internal/handlers"
	"github.com/godyy/ggs/app/game/internal/systems"
	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
)

// handleLoginCharacter 处理登录角色请求.
func handleLoginCharacter(ctx *gactor.Context, req *pbc2s.LoginCharacterReq) (*pbc2s.LoginCharacterResp, error) {
	player := handlers.GetActor[*actors.Player](ctx)

	if err := systems.Player.InitPlayer(player); err != nil {
		return nil, err
	}

	player.SetLogin()

	return &pbc2s.LoginCharacterResp{}, nil
}

// handleHearbeat 处理心跳.
func handleHeartbeat(ctx *gactor.Context, req *pbc2s.HeartbeatReq) (*pbc2s.HeartbeatResp, error) {
	player := handlers.GetActor[*actors.Player](ctx)
	player.Heartbeat()
	return &pbc2s.HeartbeatResp{}, nil
}
