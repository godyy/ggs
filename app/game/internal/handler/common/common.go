package common

import (
	"github.com/godyy/ggs/app/game/internal/app"
	"github.com/godyy/ggs/app/game/internal/handler"
	iactor "github.com/godyy/ggs/internal/infra/actor"
	actorhandler "github.com/godyy/ggs/internal/infra/actor/handler"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
	"github.com/godyy/ggskit/infra/actor"
)

func handleActorSaveResult(ctx *actor.Context, result *pbs2s.ActorSaveResult) bool {
	if result.Success {
		a, ok := ctx.Actor().Behavior().(actor.ActorWithModel)
		if !ok {
			return true
		}
		model, ok := a.GetModel().(actor.ModelDirty)
		if !ok {
			return true
		}
		model.ClearDirty()
		return true
	}

	if a, ok := ctx.Actor().Behavior().(iactor.ActorSaveWithTimer); ok {
		iactor.DelaySave(a, app.Config().Actor.SaveRetryDelay)
	}

	return true
}

func init() {
	handler.RegisterS2S(pbs2s.PID_PActorSaveResult, actorhandler.WrapCastFunc(handleActorSaveResult))
}
