package handlers

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/internal/base/consts"
	"github.com/godyy/ggs/app/internal/infra/actors"
	"github.com/godyy/ggs/internal/infra/actor"
	pbs2s "github.com/godyy/ggs/internal/proto/pb/s2s"
)

func OnActorSaveResult(ctx *gactor.Context, result *pbs2s.ActorSaveResult) {
	if result.Success {
		a, ok := ctx.Actor().Behavior().(actor.ActorWithModel)
		if !ok {
			return
		}
		model, ok := a.GetModel().(actor.ModelWithDirty)
		if !ok {
			return
		}
		model.ClearDirty()
		return
	}

	if a, ok := ctx.Actor().Behavior().(actors.ActorSaveWithTimer); ok {
		actors.DelaySave(a, consts.ActorSaveRetryDelay)
	}
}
