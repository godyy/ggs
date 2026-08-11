package systems

import (
	"reflect"
	"time"

	"github.com/godyy/ggs/app/game/internal/app"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/systems"
	"google.golang.org/protobuf/proto"
)

// broadcastTask 广播任务.
type broadcastTask struct {
	msg      proto.Message
	all      bool    // 是否广播给所有在线玩家.
	includes []int64 // 包含的玩家ID列表.
	excludes []int64 // 排除的玩家ID列表.
	ts       int64   // 时间戳, 该事件前在线的玩家才能收到广播.
}

func (bt *broadcastTask) do() {
	if bt.all {
		Player.send2AllOnlinePlayers(bt.msg, bt.excludes, bt.ts)
		return
	}
	Player.send2OnlinePlayers(bt.msg, bt.includes, bt.ts)
}

// broadcastModule 广播模块.
type broadcastModule struct {
	chTask chan *broadcastTask // 广播任务.
}

var Broadcast = systems.RegisterSystem(&broadcastModule{
	chTask: make(chan *broadcastTask, 10000),
})

func (m *broadcastModule) OnStart() {
	go m.processTask()
}

func (m *broadcastModule) OnStop() {
}

func (m *broadcastModule) send2Player(playerId int64, msg proto.Message) {
	if err := app.ActorService().Forward(actor.PlayerUID(playerId), msg); err != nil {
		logger.Get().Warnf("send msg to online player %d failed, %s:%+v, err:%v",
			playerId, reflect.TypeOf(msg).Name(), msg, err)
	}
}

func (m *broadcastModule) processTask() {
	for {
		select {
		case task := <-m.chTask:
			task.do()
		case <-systems.CStop():
			return
		}
	}
}

func (m *broadcastModule) pushTask(task *broadcastTask) {
	select {
	case m.chTask <- task:
	case <-systems.CStop():
		return
	}
}

// Send2AllOnlinePlayers 广播消息给所有在线玩家.
func (m *broadcastModule) Send2AllOnlinePlayers(msg proto.Message, excludes ...int64) {
	m.pushTask(&broadcastTask{
		msg:      msg,
		all:      true,
		excludes: excludes,
		ts:       time.Now().Unix(),
	})
}

// Send2OnlinePlayers 广播消息给指定玩家.
func (m *broadcastModule) Send2OnlinePlayers(msg proto.Message, includes ...int64) {
	m.pushTask(&broadcastTask{
		msg:      msg,
		all:      false,
		includes: includes,
		ts:       time.Now().Unix(),
	})
}
