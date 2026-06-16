package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actors"
	"github.com/godyy/ggskit/infra/actor"
)

// mdCheckLogin 检查是否已登录.
func mdCheckLogin(ctx *gactor.Context) {
	player := actor.CtxActor[*actors.Player](ctx)
	if !player.IsLogin() {
		ctx.Abort()
	}
}
