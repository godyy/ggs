package actors

import (
	"fmt"

	"github.com/godyy/gactor"
)

// lifeCycleCB 生命周期回调接口.
type lifeCycleCB interface {
	// OnStart 启动回调.
	OnStart(gactor.ActorBehavior) error
	// OnStop 停止回调.
	OnStop(gactor.ActorBehavior)
}

var lifeCycleCBs = map[uint16]lifeCycleCB{}

// registerLifeCycleCB 注册生命周期回调.
func registerLifeCycleCB(category uint16, cb lifeCycleCB) {
	if _, ok := lifeCycleCBs[category]; ok {
		panic(fmt.Errorf("life cycle cb of category %d already registered", category))
	}
	lifeCycleCBs[category] = cb
}

// LifeCycleCB 泛型生命周期回调接口, 方便实现.
type LifeCycleCB[Actor gactor.ActorBehavior] interface {
	// OnStart 启动回调.
	OnStart(Actor) error
	// OnStop 停止回调.
	OnStop(Actor)
}

// lifeCycleWrapper 生命周期回调包装器, 用于将泛型生命周期回调转换为普通生命周期回调.
type lifeCycleWrapper[Actor gactor.ActorBehavior] struct {
	cb LifeCycleCB[Actor]
}

func (w *lifeCycleWrapper[Actor]) OnStart(actor gactor.ActorBehavior) error {
	return w.cb.OnStart(actor.(Actor))
}

func (w *lifeCycleWrapper[Actor]) OnStop(actor gactor.ActorBehavior) {
	w.cb.OnStop(actor.(Actor))
}

// RegisterLifeCycleCB 泛型注册生命周期回调.
func RegisterLifeCycleCB[Actor gactor.ActorBehavior](category uint16, cb LifeCycleCB[Actor]) {
	registerLifeCycleCB(category, &lifeCycleWrapper[Actor]{cb: cb})
}

// callOnStart 调用生命周期回调的OnStart方法.
func callOnStart(actor gactor.ActorBehavior) error {
	category := actor.GetActor().ActorUID().Category
	cb, ok := lifeCycleCBs[category]
	if !ok {
		return nil
	}
	return cb.OnStart(actor)
}

// callOnStop 调用生命周期回调的OnStop方法.
func callOnStop(actor gactor.ActorBehavior) {
	category := actor.GetActor().ActorUID().Category
	cb, ok := lifeCycleCBs[category]
	if !ok {
		return
	}
	cb.OnStop(actor)
}
