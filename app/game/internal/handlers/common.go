package handlers

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/actors"
	pbs2s "github.com/godyy/ggs/internal/proto/pb/s2s"
)

func OnActorSaveResult(ctx *gactor.Context, result *pbs2s.ActorSaveResult) {
	if result.Success {
		return
	}

	actor, ok := ctx.Actor().Behavior().(actors.ActorCouldPersist)
	if !ok {
		return
	}

	actor.AsyncSaveAll()
}
