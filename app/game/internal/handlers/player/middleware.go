package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/actors"
	"github.com/godyy/ggs/app/game/internal/handlers"
)

// mdCheckLogin 检查是否已登录.
func mdCheckLogin(ctx *gactor.Context) {
	player := handlers.GetActor[*actors.Player](ctx)
	if !player.IsLogin() {
		ctx.Abort()
	}
}
