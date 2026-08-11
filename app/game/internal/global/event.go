package global

import (
	"github.com/godyy/ggs/app/game/internal/base/event"
)

// eventDispatcher 事件分发器
var eventDispatcher = event.NewDispatcher()

func NewEventWrapper[Creator any, Params any]() event.Wrapper[Creator, Params] {
	return event.NewWrapper[Creator, Params](eventDispatcher)
}

func NewEventWrapperNoParams[Creator any]() event.WrapperNoParams[Creator] {
	return event.NewWrapperNoParams[Creator](eventDispatcher)
}
