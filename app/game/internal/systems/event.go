package systems

import (
	"github.com/godyy/ggs/app/game/internal/base/event"
	"github.com/godyy/ggs/internal/infra/systems"
)

// eventModule 事件模块
type eventModule struct {
	dispatcher *event.Dispatcher
}

var Event = systems.RegisterSystem(&eventModule{
	dispatcher: event.NewDispatcher(),
})

func (m *eventModule) OnStart() {
}
func (m *eventModule) OnStop() {
}
