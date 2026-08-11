package systems

import (
	"time"

	"github.com/godyy/ggs/app/game/internal/base/event"
	"github.com/godyy/ggs/app/game/internal/global"
	"github.com/godyy/ggs/internal/infra/actor/actors"
)

func (m *playerModule) initEvent() {
	eventWrapper := global.NewEventWrapperNoParams[*actors.Player]()
	eventWrapper.ListenKind(event.KindPlayerOnline, m.onPlayerOnline, false)
	eventWrapper.ListenKind(event.KindPlayerOffline, m.onPlayerOffline, false)
}

func (m *playerModule) onPlayerOnline(e event.EventWNoParams[*actors.Player]) error {
	m.onlinePlayers.Store(e.Creator.ID(), time.Now().Unix())
	return nil
}

func (m *playerModule) onPlayerOffline(e event.EventWNoParams[*actors.Player]) error {
	m.onlinePlayers.Delete(e.Creator.ID())
	return nil
}
