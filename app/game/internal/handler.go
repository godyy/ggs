package internal

import (
	"github.com/godyy/ggs/app/game/internal/handler/player"
	"github.com/godyy/ggs/app/game/internal/handler/server"
	"github.com/godyy/ggs/internal/infra/actor"
)

// ActorHandler Actor请求处理器.
func ActorHandler(ctx *actor.Context) {
	switch actor.Category(ctx.Actor().Category()) {
	case actor.CategoryServer:
		server.Handle(ctx)
	case actor.CategoryPlayer:
		player.Handle(ctx)
	default:
		ctx.Abort()
	}
}
