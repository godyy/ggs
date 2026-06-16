package internal

import (
	"github.com/godyy/ggs/app/game/internal/handlers/player"
	"github.com/godyy/ggs/app/game/internal/handlers/server"
	"github.com/godyy/ggs/internal/infra/actors"
	"github.com/godyy/ggskit/infra/actor"
)

// ActorHandler Actor请求处理器.
func ActorHandler(ctx *actor.Context) {
	switch actors.Category(ctx.Actor().Category()) {
	case actors.CategoryServer:
		server.Handle(ctx)
	case actors.CategoryPlayer:
		player.Handle(ctx)
	default:
		ctx.Abort()
	}
}
