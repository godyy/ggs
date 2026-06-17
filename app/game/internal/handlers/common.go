package handlers

import (
	"github.com/godyy/ggs/internal/base/consts"
	iactor "github.com/godyy/ggs/internal/infra/actor"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
	"github.com/godyy/ggskit/infra/actor"
)

func OnActorSaveResult(ctx *actor.Context, result *pbs2s.ActorSaveResult) {
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

	if a, ok := ctx.Actor().Behavior().(iactor.ActorSaveWithTimer); ok {
		iactor.DelaySave(a, consts.ActorSaveRetryDelay)
	}
}
