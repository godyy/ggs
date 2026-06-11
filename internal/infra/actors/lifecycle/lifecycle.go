package lifecycle

import (
	"fmt"

	"github.com/godyy/ggskit/infra/actor"
)

// handler 生命周期回调处理器.
type handler interface {
	// OnStart OnStart 回调.
	OnStart(actor.Actor) error

	// OnStop OnStop 回调.
	OnStop(actor.Actor)
}

// cHandler CActor 生命周期回调处理器.
type cHandler interface {
	handler

	// OnConnected OnConnected 回调.
	OnConnected(actor.CActor)

	// OnDisconnected OnDisconnected 回调.
	OnDisconnected(actor.CActor)
}

var handlers = map[actor.Category]handler{}

// registerHandler 注册生命周期回调处理器.
func registerHandler(category actor.Category, cb handler) {
	if _, ok := handlers[category]; ok {
		panic(fmt.Errorf("lifecycle handler of category %v already registered", category))
	}
	handlers[category] = cb
}

// getHandler 获取生命周期回调处理器.
func getHandler(category actor.Category) handler {
	return handlers[category]
}

// Handler 生命周期回调处理器泛型封装.
type Handler[Actor actor.Actor] interface {
	// OnStart OnStart 回调.
	OnStart(Actor) error

	// OnStop OnStop 回调.
	OnStop(Actor)
}

// CHandler CActor 生命周期回调处理器泛型封装.
type CHandler[Actor actor.CActor] interface {
	Handler[Actor]

	// OnConnected OnConnected 回调.
	OnConnected(Actor)

	// OnDisconnected OnDisconnected 回调.
	OnDisconnected(Actor)
}

// hanndlerWrapper 生命周期回调处理器泛型包装器
type hanndlerWrapper[Actor actor.Actor] struct {
	h Handler[Actor]
}

func (w *hanndlerWrapper[Actor]) OnStart(actor actor.Actor) error {
	return w.h.OnStart(actor.(Actor))
}

func (w *hanndlerWrapper[Actor]) OnStop(actor actor.Actor) {
	w.h.OnStop(actor.(Actor))
}

// cHanndlerWrapper CActor 生命周期回调处理器泛型包装器
type cHanndlerWrapper[Actor actor.CActor] struct {
	h CHandler[Actor]
}

func (w *cHanndlerWrapper[Actor]) OnStart(actor actor.Actor) error {
	return w.h.OnStart(actor.(Actor))
}

func (w *cHanndlerWrapper[Actor]) OnStop(actor actor.Actor) {
	w.h.OnStop(actor.(Actor))
}

func (w *cHanndlerWrapper[Actor]) OnConnected(actor actor.CActor) {
	w.h.OnConnected(actor.(Actor))
}

func (w *cHanndlerWrapper[Actor]) OnDisconnected(actor actor.CActor) {
	w.h.OnDisconnected(actor.(Actor))
}

// RegisterHandler 注册生命周期回调处理器泛型封装.
func RegisterHandler[Actor actor.Actor](category actor.Category, h Handler[Actor]) {
	registerHandler(category, &hanndlerWrapper[Actor]{h: h})
}

// RegisterCHandler 注册CActor生命周期回调处理器泛型封装.
func RegisterCHandler[Actor actor.CActor](category actor.Category, cb CHandler[Actor]) {
	registerHandler(category, &cHanndlerWrapper[Actor]{h: cb})
}

// OnStart 调用OnStart回调.
func OnStart(a actor.Actor) error {
	category := actor.Category(a.ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return nil
	}
	return handler.OnStart(a)
}

// OnStop 调用OnStop回调.
func OnStop(a actor.Actor) {
	category := actor.Category(a.ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return
	}
	handler.OnStop(a)
}

// OnConnected 调用OnConnected回调.
func OnConnected(a actor.CActor) {
	category := actor.Category(a.ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return
	}
	handler.(cHandler).OnConnected(a)
}

// OnDisconnected 调用OnDisconnected回调.
func OnDisconnected(a actor.CActor) {
	category := actor.Category(a.ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return
	}
	handler.(cHandler).OnDisconnected(a)
}
