package persist

import (
	"errors"
	"math/rand/v2"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggskit/infra/actor"
	"github.com/godyy/ggskit/infra/mongobd"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

const (
	// delayRandom 持久化定时器随机延迟时间.
	delayRandom = 5 * time.Second

	// retryDelay 持久化重试延迟时间.
	retryDelay = 5 * time.Second
)

// BD 持久化后台.
type BD interface {
	Exec(hashKey any, op mongobd.Op) error
	Add(hashKey any, op mongobd.Op, callback func(mongobd.Op)) error
}

// LoadModel 同步加载Model.
func LoadModel(m actor.Model, db string) (exists bool, err error) {
	checkState()
	op := mongobd.NewOp[mongobd.OpLoad](db, m.GetCollection()).
		SetFilter(m.GetFilter()).
		SetPrimary(true).
		SetTarget(m)
	for i := range persistRetryDelays {
		persistRetrySleep(i)
		if err = bd.Exec(m.GetHashKey(), op); err == nil || !mongo.IsNetworkError(err) {
			break
		}
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return true, nil
}

// extractModelDirty 提取Model的脏数据.
// 若model实现了ModelWithDirty接口,则根据是否全量更新来准备更新数据.
// 若model未实现ModelWithDirty接口,则直接全量更新.
func extractModelDirty(m actor.Model) (update []byte, upsert bool, err error) {
	if modelDirty, ok := m.(actor.ModelDirty); ok {
		// 若model实现了ModelWithDirty接口,则根据是否全量更新来准备更新数据.
		dirty, all := modelDirty.IsDirty()
		if !dirty {
			return
		}

		if all {
			// 全量更新
			update, err = bson.Marshal(m)
			upsert = true
		} else {
			// 脏数据更新
			update, err = modelDirty.MarshalBSONDirty()
		}
	} else {
		// 若model未实现ModelWithDirty接口,则直接全量更新.
		update, err = bson.Marshal(m)
		upsert = true
	}

	return
}

// clearModelDirty 清理Model的脏数据.
// 若model实现了ModelWithDirty接口,则调用ClearDirty方法清理脏数据.
func clearModelDirty(m actor.Model) {
	if modelDirty, ok := m.(actor.ModelDirty); ok {
		modelDirty.ClearDirty()
	}
}

// SaveModel 同步保存Model.
func SaveModel(m actor.Model, db string) (err error) {
	checkState()

	update, upsert, err := extractModelDirty(m)
	if err != nil {
		return err
	}

	// 执行操作符
	op := mongobd.NewOp[mongobd.OpUpdate](db, m.GetCollection()).
		SetFilter(m.GetFilter()).
		SetUpdate(bson.Raw(update)).
		SetUpsert(upsert)
	for i := range persistRetryDelays {
		persistRetrySleep(i)
		if err = bd.Exec(m.GetHashKey(), op); err == nil {
			break
		}
	}

	// 清理脏数据
	if err == nil {
		clearModelDirty(m)
	}

	return
}

// AsyncSaveModelCallback 异步保存模型回调.
type AsyncSaveModelCallback func(uid gactor.ActorUID, err error)

// AsyncSaveModel 异步保存Model.
func AsyncSaveModel(uid gactor.ActorUID, m actor.Model, db string, callback AsyncSaveModelCallback) (err error) {
	checkState()

	update, upsert, err := extractModelDirty(m)
	if err != nil {
		return err
	}

	// 添加操作符.
	op := mongobd.NewOp[mongobd.OpUpdate](db, m.GetCollection()).
		SetFilter(m.GetFilter()).
		SetUpdate(bson.Raw(update)).
		SetUpsert(upsert)
	return bd.Add(m.GetHashKey(), op, func(o mongobd.Op) {
		if o.Err() == nil {
			callback(uid, nil)
			return
		}

		logger.Get().ErrorFields("persist op async exec failed",
			zap.Uint16("category", uid.Category),
			zap.Int64("id", uid.ID),
			zap.NamedError("error", o.Err()),
		)

		callback(uid, o.Err())
	})
}

// ActorSaveWithTimer 可结合定时器实现持久化的Actor.
type ActorSaveWithTimer interface {
	actor.ActorWithModel
	SaveTimerId() gactor.TimerId
	SetSaveTimerId(gactor.TimerId)
}

// DelaySaveActorModel 延迟保存.
// 启动延迟保存timer, 等待timer触发进行保存操作.
func DelaySaveActorModel(a ActorSaveWithTimer, db string, delay time.Duration, callback AsyncSaveModelCallback) {
	checkState()
	delay = delay + time.Duration(rand.Int64N(int64(delayRandom)))
	startSaveTimer(a, db, delay, callback)
}

type saveTimerArgs struct {
	db       string
	callback AsyncSaveModelCallback
}

// startSaveTimer 启动保存timer.
func startSaveTimer(a ActorSaveWithTimer, db string, delay time.Duration, callback AsyncSaveModelCallback) {
	if a.SaveTimerId() != gactor.TimerIdNone {
		return
	}

	timerId := a.StartTimer(delay, false, &saveTimerArgs{db, callback}, onSaveTimer)
	a.SetSaveTimerId(timerId)
}

// onSaveTimer 保存定时器回调
func onSaveTimer(args *gactor.ActorTimerArgs) {
	a := args.Actor.Behavior().(ActorSaveWithTimer)
	saveArgs := args.Args.(*saveTimerArgs)
	if timerId := a.SaveTimerId(); timerId == args.TID {
		a.SetSaveTimerId(gactor.TimerIdNone)
	}

	if err := AsyncSaveModel(a.ActorUID(), a.GetModel(), saveArgs.db, saveArgs.callback); err != nil {
		uid := a.ActorUID()
		logger.Get().ErrorFields("persist async failed",
			zap.Uint16("category", uid.Category),
			zap.Int64("id", uid.ID),
			zap.NamedError("error", err),
		)

		// 尝试重新持久化
		retryDelay := retryDelay + time.Duration(rand.Int64N(int64(delayRandom)))
		startSaveTimer(a, saveArgs.db, retryDelay, saveArgs.callback)
	}
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
