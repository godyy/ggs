package actor

import (
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor/persist"
	"go.uber.org/zap"
)

type ActorSaveWithTimer = persist.ActorSaveWithTimer

type AsyncSaveCallback func(uid gactor.ActorUID, err error)

// persistor 持久化辅助结构
type persistor struct {
	saveTimerId TimerId // save 定时器ID
}

func (p *persistor) SaveTimerId() TimerId {
	return p.saveTimerId
}

func (p *persistor) SetSaveTimerId(timerId TimerId) {
	p.saveTimerId = timerId
}

// LoadModel 加载Actor模型.
func LoadModel(a ActorSaveWithTimer) (bool, error) {
	checkState()
	return persist.LoadModel(a.GetModel(), db)
}

// SaveModel 保存Actor模型.
func SaveModel(a ActorSaveWithTimer) error {
	checkState()
	return persist.SaveModel(a.GetModel(), db)
}

// AsyncSaveModel 异步保存Actor模型.
func AsyncSaveModel(a ActorSaveWithTimer) error {
	checkState()
	return persist.AsyncSaveModel(a.ActorUID(), a.GetModel(), db, asyncSaveModelCallback)
}

// DelaySave 延迟保存Actor.
func DelaySave(a ActorSaveWithTimer, delay time.Duration) {
	checkState()
	persist.DelaySaveActorModel(a, db, delay, asyncSaveModelCallback)
}

func asyncSaveModelCallback(uid gactor.ActorUID, err error) {
	if err != nil {
		logger.Get().ErrorFields("async save actor model failed, uid: %v, err: %v",
			zap.String("category", Category(uid.Category).String()),
			zap.Int64("id", uid.ID),
			zap.NamedError("error", err),
		)
	}

	asyncSaveCallback(uid, err)
}
