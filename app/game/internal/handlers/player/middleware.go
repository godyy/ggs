package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/handlers"
	"github.com/godyy/ggs/internal/infra/actors"
)

// mdCheckLogin 检查是否已登录.
func mdCheckLogin(ctx *gactor.Context) {
	player := handlers.GetActor[*actors.Player](ctx)
	if !player.IsLogin() {
		ctx.Abort()
	}
}
