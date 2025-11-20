package actors

import (
	"context"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/app"
	"github.com/godyy/ggs/app/game/internal/base/consts"
	"github.com/godyy/ggs/app/internal/actors"
	"github.com/godyy/ggs/internal/base/actor"
	"github.com/godyy/ggs/internal/libs/logger"
	pbs2s "github.com/godyy/ggs/internal/proto/pb/s2s"
	"go.uber.org/zap"
)

type ActorSaveWithTimer = actors.ActorSaveWithTimer

// persistor 持久化辅助结构
type persistor struct {
	saveTimerId gactor.TimerId // save 定时器ID
}

func (p *persistor) SaveTimerId() gactor.TimerId {
	return p.saveTimerId
}

func (p *persistor) SetSaveTimerId(timerId gactor.TimerId) {
	p.saveTimerId = timerId
}

// LoadActorModel 加载Actor模型.
func LoadActorModel(a ActorSaveWithTimer) error {
	return actors.LoadModel(a.GetModel(), app.Env().DB(), app.MongoBD())
}

// SaveActorModel 保存Actor模型.
func SaveActorModel(a ActorSaveWithTimer) error {
	return actors.SaveModel(a.ActorUID(), a.GetModel(), app.Env().DB(), app.MongoBD())
}

// AsyncSaveActorModel 异步保存Actor模型.
func AsyncSaveActorModel(a ActorSaveWithTimer) error {
	return actors.AsyncSaveModel(a.ActorUID(), a.GetModel(), app.Env().DB(), app.MongoBD(), asyncSaveActorModelCallback)
}

// DelaySaveActor 延迟保存Actor.
func DelaySaveActor(a ActorSaveWithTimer, delay time.Duration) {
	actors.DelaySaveActorModel(a, app.Env().DB(), app.MongoBD(), delay, asyncSaveActorModelCallback)
}

func asyncSaveActorModelCallback(uid gactor.ActorUID, err error) {
	if err != nil {
		logger.GetLogger().ErrorFields("async save actor model failed, uid: %v, err: %v",
			zap.String("category", actor.Category(uid.Category).String()),
			zap.Int64("id", uid.ID),
			zap.NamedError("error", err),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), consts.ActorCastTimeout)
	defer cancel()
	if err := app.ActorService().Cast(ctx, uid, &pbs2s.ActorSaveResult{
		Success: err == nil,
	}); err != nil {
		logger.GetLogger().ErrorFields("cast persist result to actor",
			zap.String("category", actor.Category(uid.Category).String()),
			zap.Int64("id", uid.ID),
			zap.NamedError("error", err),
		)
	}
}
