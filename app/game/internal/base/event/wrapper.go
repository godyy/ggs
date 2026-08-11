package event

import (
	"github.com/godyy/gevent"
)

// EventW 包装后的事件
type EventW[Creator, Params any] = gevent.EventW[Kind, int64, Creator, Params]

// Wrapper 事件包装器
type Wrapper[Creator, Params any] struct {
	gevent.Wrapper[Kind, int64, Creator, Params]
}

func NewWrapper[Creator, Params any](d *Dispatcher) Wrapper[Creator, Params] {
	return Wrapper[Creator, Params]{
		Wrapper: gevent.Wrapper[Kind, int64, Creator, Params]{
			Dispatcher: d.Dispatcher,
		},
	}
}

// ListenKind 监听类型事件
func (w Wrapper[Creator, Params]) ListenKind(kind Kind, cb func(EventW[Creator, Params]) error, once bool) LID {
	return w.Wrapper.ListenKind(kind, cb, once)
}

// ListenKindOnce 监听类型事件（仅执行一次）
func (w Wrapper[Creator, Params]) ListenKindOnce(kind Kind, cb func(EventW[Creator, Params]) error) LID {
	return w.ListenKind(kind, cb, true)
}

// Listen 监听事件
func (w Wrapper[Creator, Params]) Listen(id ID, cb func(EventW[Creator, Params]) error, once bool) LID {
	return w.Wrapper.Listen(id, cb, once)
}

// ListenOnce 监听事件（仅执行一次）
func (w Wrapper[Creator, Params]) ListenOnce(id ID, cb func(EventW[Creator, Params]) error) LID {
	return w.Listen(id, cb, true)
}

// DispatchKind 分发类型事件
func (w Wrapper[Creator, Params]) DispatchKind(kind Kind, creator Creator, params Params) error {
	return w.Dispatch(MakeKindID(kind), creator, params)
}

// EventWNoParams 包装后的事件（无参数）
type EventWNoParams[Creator any] = gevent.EventW[Kind, int64, Creator, any]

// WrapperNoParams 事件包装器（无参数）
type WrapperNoParams[Creator any] struct {
	Wrapper[Creator, any]
}

func NewWrapperNoParams[Creator any](d *Dispatcher) WrapperNoParams[Creator] {
	return WrapperNoParams[Creator]{
		Wrapper: NewWrapper[Creator, any](d),
	}
}

// ListenKind 监听类型事件
func (w WrapperNoParams[Creator]) ListenKind(kind Kind, cb func(EventWNoParams[Creator]) error, once bool) LID {
	return w.Wrapper.ListenKind(kind, cb, once)
}

// ListenKindOnce 监听类型事件（仅执行一次）
func (w WrapperNoParams[Creator]) ListenKindOnce(kind Kind, cb func(EventWNoParams[Creator]) error) LID {
	return w.ListenKind(kind, cb, true)
}

// Listen 监听事件
func (w WrapperNoParams[Creator]) Listen(id ID, cb func(EventWNoParams[Creator]) error, once bool) LID {
	return w.Wrapper.Listen(id, cb, once)
}

// ListenOnce 监听事件（仅执行一次）
func (w WrapperNoParams[Creator]) ListenOnce(id ID, cb func(EventWNoParams[Creator]) error) LID {
	return w.Listen(id, cb, true)
}

// Dispatch 分发事件
func (w WrapperNoParams[Creator]) Dispatch(id ID, creator Creator) error {
	return w.Wrapper.Dispatch(id, creator, nil)
}

// DispatchKind 分发类型事件
func (w WrapperNoParams[Creator]) DispatchKind(kind Kind, creator Creator) error {
	return w.Wrapper.DispatchKind(kind, creator, nil)
}
