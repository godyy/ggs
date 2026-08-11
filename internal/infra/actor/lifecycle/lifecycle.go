package lifecycle

import (
	"fmt"

	"github.com/godyy/ggs/internal/infra/actor"
)

// handler 生命周期回调处理器.
type handler interface {
	// OnActorStart Actor OnStart 回调.
	OnActorStart(actor.Actor) error

	// OnActorStop Actor OnStop 回调.
	OnActorStop(actor.Actor)
}

// cHandler CActor 生命周期回调处理器.
type cHandler interface {
	handler

	// OnActorConnected CActor OnConnected 回调.
	OnActorConnected(actor.CActor)

	// OnActorDisconnected CActor OnDisconnected 回调.
	OnActorDisconnected(actor.CActor)
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
	// OnActorStart Actor OnStart 回调.
	OnActorStart(Actor) error

	// OnActorStop Actor OnStop 回调.
	OnActorStop(Actor)
}

// CHandler CActor 生命周期回调处理器泛型封装.
type CHandler[Actor actor.CActor] interface {
	Handler[Actor]

	// OnActorConnected CActor OnConnected 回调.
	OnActorConnected(Actor)

	// OnActorDisconnected CActor OnDisconnected 回调.
	OnActorDisconnected(Actor)
}

// hanndlerWrapper 生命周期回调处理器泛型包装器
type hanndlerWrapper[Actor actor.Actor] struct {
	h Handler[Actor]
}

func (w *hanndlerWrapper[Actor]) OnActorStart(actor actor.Actor) error {
	return w.h.OnActorStart(actor.(Actor))
}

func (w *hanndlerWrapper[Actor]) OnActorStop(actor actor.Actor) {
	w.h.OnActorStop(actor.(Actor))
}

// cHanndlerWrapper CActor 生命周期回调处理器泛型包装器
type cHanndlerWrapper[Actor actor.CActor] struct {
	h CHandler[Actor]
}

func (w *cHanndlerWrapper[Actor]) OnActorStart(actor actor.Actor) error {
	return w.h.OnActorStart(actor.(Actor))
}

func (w *cHanndlerWrapper[Actor]) OnActorStop(actor actor.Actor) {
	w.h.OnActorStop(actor.(Actor))
}

func (w *cHanndlerWrapper[Actor]) OnActorConnected(actor actor.CActor) {
	w.h.OnActorConnected(actor.(Actor))
}

func (w *cHanndlerWrapper[Actor]) OnActorDisconnected(actor actor.CActor) {
	w.h.OnActorDisconnected(actor.(Actor))
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
	return handler.OnActorStart(a)
}

// OnStop 调用OnStop回调.
func OnStop(a actor.Actor) {
	category := actor.Category(a.ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return
	}
	handler.OnActorStop(a)
}

// OnConnected 调用OnConnected回调.
func OnConnected(a actor.CActor) {
	category := actor.Category(a.ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return
	}
	handler.(cHandler).OnActorConnected(a)
}

// OnDisconnected 调用OnDisconnected回调.
func OnDisconnected(a actor.CActor) {
	category := actor.Category(a.ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return
	}
	handler.(cHandler).OnActorDisconnected(a)
}
