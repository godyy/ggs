package persist

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	pbs2s "github.com/godyy/ggs/internal/proto/pb/s2s"

	"github.com/godyy/ggs/app/game/internal/actors/model"
	"github.com/godyy/ggs/app/game/internal/consts"
	"github.com/godyy/ggs/app/game/internal/env"

	"github.com/godyy/ggs/internal/core/actor"
	"github.com/godyy/ggs/internal/core/db/mongobd"

	"github.com/godyy/gactor"
	mactor "github.com/godyy/ggs/app/game/internal/modules/actor"
	mmongobd "github.com/godyy/ggs/app/game/internal/modules/mongobd"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

const (
	// delayRandom 持久化定时器随机延迟时间.
	delayRandom = 5 * time.Second

	// retryDelay 持久化重试延迟时间.
	retryDelay = 5 * time.Second

	castTimeout = 1 * time.Second
)

type Actor interface {
	gactor.Actor
	GetModel() model.Model
	SaveTimerId() gactor.TimerId
	SetSaveTimerId(gactor.TimerId)
}

// DelaySaveActor 延迟保存.
// 启动延迟保存timer, 等待timer触发进行保存操作.
func DelaySaveActor(a Actor) {
	delay := consts.DirtyPersistDelay + time.Duration(rand.Int64N(int64(delayRandom)))
	startSaveTimer(a, delay)
}

// startSaveTimer 启动保存timer.
func startSaveTimer(a Actor, delay time.Duration) {
	if a.SaveTimerId() != gactor.TimerIdNone {
		return
	}

	timerId := a.StartTimer(delay, false, nil, onSaveTimer)
	a.SetSaveTimerId(timerId)
}

// onSaveTimer 保存定时器回调
func onSaveTimer(args *gactor.ActorTimerArgs) {
	a := args.Actor.Behavior().(Actor)
	if timerId := a.SaveTimerId(); timerId == args.TID {
		a.SetSaveTimerId(gactor.TimerIdNone)
	}
	if err := saveModel(a.ActorUID(), a.GetModel(), true); err != nil {
		uid := a.ActorUID()
		logger.ErrorFields("persist sync failed",
			zap.String("category", actor.CategoryName(uid.Category)),
			zap.Int64("id", uid.ID),
			zap.NamedError("error", err),
		)
		// 尝试重新持久化
		startSaveTimer(a, retryDelay)
	}
}

// SaveModel 同步保存Model.
func SaveModel(uid gactor.ActorUID, m model.Model) (err error) {
	return saveModel(uid, m, false)
}

// saveModel 保存Model.
// 支持异步保存, 若async为true,则会将操作符添加到异步队列中, 否则会阻塞等待操作完成.
func saveModel(uid gactor.ActorUID, m model.Model, async bool) (err error) {
	// 准备需要更新的数据.
	var (
		update []byte
		upsert bool
	)
	if modelDirty, ok := m.(model.ModelWithDirty); ok {
		// 若model实现了ModelWithDirty接口,则根据是否全量更新来准备更新数据.
		dirty, all := modelDirty.IsDirty()
		if !dirty {
			return nil
		}

		if all {
			// 全量更新
			update, err = bson.Marshal(m)
			upsert = true
		} else {
			// 脏数据更新
			update, err = modelDirty.MarshalBSONDirty()
			if err == nil {
				defer func() {
					if err == nil {
						modelDirty.ClearDirty()
					}
				}()
			}
		}
	} else {
		// 若model未实现ModelWithDirty接口,则直接全量更新.
		update, err = bson.Marshal(m)
		upsert = true
	}

	if err != nil {
		return
	}

	// 添加操作符
	op := mongobd.NewOp[mongobd.OpUpdate](env.All().DB(), m.GetCollection()).
		SetFilter(m.GetFilter()).
		SetUpdate(bson.Raw(update)).
		SetUpsert(upsert)
	if async {
		if err = mmongobd.Add(m.GetHashKey(), op, func(o mongobd.Op) {
			if o.Err() == nil {
				return
			}

			logger.ErrorFields("persist op async exec failed",
				zap.String("category", actor.CategoryName(uid.Category)),
				zap.Int64("id", uid.ID),
				zap.NamedError("error", o.Err()),
			)

			// 持久化失败后的处理, 发送消息通知actor
			ctx, cancel := context.WithTimeout(context.Background(), castTimeout)
			defer cancel()
			if err := mactor.Cast(ctx, uid, &pbs2s.ActorSaveResult{
				Success: false,
			}); err != nil {
				logger.ErrorFields("cast persist failed to actor",
					zap.String("category", actor.CategoryName(uid.Category)),
					zap.Int64("id", uid.ID),
					zap.NamedError("error", err),
				)
			}

		}); err != nil {
			return err
		}
	} else {
		for i := range persistRetryDelays {
			persistRetrySleep(i)
			if err = mmongobd.Exec(m.GetHashKey(), op); err == nil {
				break
			}
		}

	}

	return nil
}

// LoadModel 同步加载Model.
func LoadModel(m model.Model) (err error) {
	op := mongobd.NewOp[mongobd.OpLoad](env.All().DB(), m.GetCollection()).
		SetFilter(m.GetFilter()).
		SetPrimary(true).
		SetTarget(m)
	for i := range persistRetryDelays {
		persistRetrySleep(i)
		if err = mmongobd.Exec(m.GetHashKey(), op); err == nil || !mongo.IsNetworkError(err) {
			break
		}
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	return
}

// maxPersistRetries 最大持久化重试次数.
const maxPersistRetries = 3

// persistRetryDelays 持久化重试延迟时间.
var persistRetryDelays = [1 + maxPersistRetries]time.Duration{
	0,
	200 * time.Millisecond,
	500 * time.Millisecond,
	1000 * time.Millisecond,
}

// persistRetrySleep 持久化重试延迟.
func persistRetrySleep(retry int) {
	delay := time.Duration(0)
	if retry < len(persistRetryDelays) {
		delay = persistRetryDelays[retry]
	} else {
		delay = persistRetryDelays[len(persistRetryDelays)-1]
	}
	if delay > 0 {
		time.Sleep(delay)
	}
}
