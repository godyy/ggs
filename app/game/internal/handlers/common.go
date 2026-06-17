package handlers

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/infra/actors"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
	"github.com/godyy/ggskit/infra/actor"
)

func OnActorSaveResult(ctx *gactor.Context, result *pbs2s.ActorSaveResult) {
	if result.Success {
		a, ok := ctx.Actor().Behavior().(actor.ActorWithModel)
		if !ok {
			return
		}
		model, ok := a.GetModel().(actor.ModelDirty)
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
