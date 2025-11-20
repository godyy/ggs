package handlers

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/actors"
	"github.com/godyy/ggs/app/game/internal/base/consts"
	"github.com/godyy/ggs/internal/base/actor"
	"github.com/godyy/ggs/internal/base/actor/model"
	pbs2s "github.com/godyy/ggs/internal/proto/pb/s2s"
)

func OnActorSaveResult(ctx *gactor.Context, result *pbs2s.ActorSaveResult) {
	if result.Success {
		actor, ok := ctx.Actor().Behavior().(actor.ActorWithModel)
		if !ok {
			return
		}
		model, ok := actor.GetModel().(model.ModelWithDirty)
		if !ok {
			return
		}
		model.ClearDirty()
		return
	}

	if actor, ok := ctx.Actor().Behavior().(actors.ActorSaveWithTimer); ok {
		actors.DelaySaveActor(actor, consts.ActorSaveRetryDelay)
	}
}
