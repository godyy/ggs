package player

import (
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
)

// mdCheckLogin 检查是否已登录.
func mdCheckLogin(ctx *actor.Context) {
	player := actor.CtxActor[*actors.Player](ctx)
	if !player.IsLogin() {
		ctx.Abort()
	}
}
