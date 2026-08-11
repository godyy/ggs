package event

import "github.com/godyy/gevent"

// Dispatcher 事件分发器
type Dispatcher struct {
	*gevent.Dispatcher[Kind, int64]
}

// NewDispatcher 创建事件分发器
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		Dispatcher: gevent.NewDispatcher[Kind, int64](),
	}
}

// DispatchKind 分发类型事件
func (d *Dispatcher) DispatchKind(kind Kind, creator, params any) error {
	return d.Dispatcher.Dispatch(MakeKindID(kind), creator, params)
}
