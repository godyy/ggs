package lifecycle

import (
	"fmt"

	"github.com/godyy/ggs/internal/infra/actor"
)

// handler 生命周期回调处理器.
type handler interface {
	// OnActorStart OnStart 回调.
	OnActorStart(actor.ActorBehavior) error

	// OnActorStop OnStop 回调.
	OnActorStop(actor.ActorBehavior)
}

// cHandler CActor 生命周期回调处理器.
type cHandler interface {
	handler

	// OnActorConnected OnConnected 回调.
	OnActorConnected(actor.CActorBehavior)

	// OnActorDisconnected OnDisconnected 回调.
	OnActorDisconnected(actor.CActorBehavior)
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

// Handler ActorBehavior 生命周期回调处理器泛型封装.
type Handler[Actor actor.ActorBehavior] interface {
	// OnActorStart OnStart 回调.
	OnActorStart(Actor) error

	// OnActorStop OnStop 回调.
	OnActorStop(Actor)
}

// CHandler CActorBehavior 生命周期回调处理器泛型封装.
type CHandler[Actor actor.CActorBehavior] interface {
	Handler[Actor]

	// OnActorConnected OnConnected 回调.
	OnActorConnected(Actor)

	// OnActorDisconnected OnDisconnected 回调.
	OnActorDisconnected(Actor)
}

// hanndlerWrapper ActorBehavior 生命周期回调处理器泛型包装器
type hanndlerWrapper[Actor actor.ActorBehavior] struct {
	h Handler[Actor]
}

func (w *hanndlerWrapper[Actor]) OnActorStart(actor actor.ActorBehavior) error {
	return w.h.OnActorStart(actor.(Actor))
}

func (w *hanndlerWrapper[Actor]) OnActorStop(actor actor.ActorBehavior) {
	w.h.OnActorStop(actor.(Actor))
}

// cHanndlerWrapper CActorBehavior 生命周期回调处理器泛型包装器
type cHanndlerWrapper[Actor actor.CActorBehavior] struct {
	h CHandler[Actor]
}

func (w *cHanndlerWrapper[Actor]) OnActorStart(actor actor.ActorBehavior) error {
	return w.h.OnActorStart(actor.(Actor))
}

func (w *cHanndlerWrapper[Actor]) OnActorStop(actor actor.ActorBehavior) {
	w.h.OnActorStop(actor.(Actor))
}

func (w *cHanndlerWrapper[Actor]) OnActorConnected(actor actor.CActorBehavior) {
	w.h.OnActorConnected(actor.(Actor))
}

func (w *cHanndlerWrapper[Actor]) OnActorDisconnected(actor actor.CActorBehavior) {
	w.h.OnActorDisconnected(actor.(Actor))
}

// RegisterHandler 注册ActorBehavior生命周期回调处理器泛型封装.
func RegisterHandler[Actor actor.ActorBehavior](category actor.Category, h Handler[Actor]) {
	registerHandler(category, &hanndlerWrapper[Actor]{h: h})
}

// RegisterCHandler 注册CActorBehavior生命周期回调处理器泛型封装.
func RegisterCHandler[Actor actor.CActorBehavior](category actor.Category, cb CHandler[Actor]) {
	registerHandler(category, &cHanndlerWrapper[Actor]{h: cb})
}

// OnStart 调用OnStart回调.
func OnStart(a actor.ActorBehavior) error {
	category := actor.Category(a.GetActor().ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return nil
	}
	return handler.OnActorStart(a)
}

// OnStop 调用OnStop回调.
func OnStop(a actor.ActorBehavior) {
	category := actor.Category(a.GetActor().ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return
	}
	handler.OnActorStop(a)
}

// OnConnected 调用OnConnected回调.
func OnConnected(a actor.CActorBehavior) {
	category := actor.Category(a.GetActor().ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return
	}
	handler.(cHandler).OnActorConnected(a)
}

// OnDisconnected 调用OnDisconnected回调.
func OnDisconnected(a actor.CActorBehavior) {
	category := actor.Category(a.GetActor().ActorUID().Category)
	handler := getHandler(category)
	if handler == nil {
		return
	}
	handler.(cHandler).OnActorDisconnected(a)
}
