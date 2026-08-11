package systems

import (
	"fmt"
	"reflect"
	"slices"
	"sync"

	"github.com/godyy/ggs/app/game/internal/app"
	"github.com/godyy/ggs/app/game/internal/base/event"
	"github.com/godyy/ggs/app/game/internal/global"
	"github.com/godyy/ggs/app/game/internal/handler"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/lifecycle"
	"github.com/godyy/ggs/internal/infra/actor/model/player"
	"github.com/godyy/ggs/internal/infra/systems"
	"google.golang.org/protobuf/proto"
)

type playerModule struct {
	onlinePlayers *sync.Map // 在线玩家ID列表.
}

var Player = systems.RegisterSystem(&playerModule{
	onlinePlayers: &sync.Map{},
})

func init() {
	lifecycle.RegisterCHandler(actor.CategoryPlayer, Player)
}

func (m *playerModule) OnStart() {
	m.initEvent()
}

func (m *playerModule) OnStop() {
}

// OnActorStart Player OnStart回调.
func (m *playerModule) OnActorStart(p *actors.Player) error {
	return nil
}

// OnActorStop Player OnStop回调.
func (m *playerModule) OnActorStop(p *actors.Player) {

}

// OnActorConnected Player OnConnected回调.
func (m *playerModule) OnActorConnected(p *actors.Player) {

}

// OnActorDisconnected Player OnDisconnected回调.
func (m *playerModule) OnActorDisconnected(p *actors.Player) {
	if err := global.NewEventWrapperNoParams[*actors.Player]().DispatchKind(event.KindPlayerOffline, p); err != nil {
		handler.Logger().Errorf("dispatch player offline event, err:%v", err)
		return
	}
}

// InitPlayer 初始化player.
func (m *playerModule) InitPlayer(p *actors.Player) error {
	if p.Model.IsInit() {
		return nil
	}

	base := actor.GetActorModule[*player.BaseInfo](p, true)
	base.Name = fmt.Sprintf("player%d", p.ID())
	Items.init(p)
	p.Model.Version = consts.VersionInit
	p.SetAllDirty()

	return nil
}

func (m *playerModule) IsPlayerOnline(id int64) bool {
	_, ok := m.onlinePlayers.Load(id)
	return ok
}

func (m *playerModule) Send2OnlinePlayer(playerId int64, msg proto.Message) {
	if !m.IsPlayerOnline(playerId) {
		return
	}
	if err := app.ActorService().Forward(actor.PlayerUID(playerId), msg); err != nil {
		logger.Get().Warnf("send msg to online player %d failed, %s:%+v, err:%v",
			playerId, reflect.TypeOf(msg).Name(), msg, err)
	}
}

func (m *playerModule) send2Player(playerId int64, msg proto.Message) {
	if err := app.ActorService().Forward(actor.PlayerUID(playerId), msg); err != nil {
		logger.Get().Warnf("send msg to online player %d failed, %s:%+v, err:%v",
			playerId, reflect.TypeOf(msg).Name(), msg, err)
	}
}

func (m *playerModule) send2AllOnlinePlayers(msg proto.Message, excludes []int64, ts int64) {
	slices.Sort(excludes)
	Player.onlinePlayers.Range(func(key, value any) bool {
		playerId := key.(int64)
		loginTs := value.(int64)
		if loginTs <= ts && slices.Contains(excludes, playerId) {
			return true
		}
		m.send2Player(playerId, msg)
		return true
	})
}

func (m *playerModule) send2OnlinePlayers(msg proto.Message, includes []int64, ts int64) {
	for _, playerId := range includes {
		value, ok := Player.onlinePlayers.Load(playerId)
		if !ok {
			continue
		}
		loginTs := value.(int64)
		if loginTs > ts {
			continue
		}
		m.send2Player(playerId, msg)
	}
}
