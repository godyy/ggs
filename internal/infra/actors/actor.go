package actors

import (
	"context"
	"time"

	"github.com/godyy/ggskit/infra/actor"
	"google.golang.org/protobuf/proto"
)

const (
	ActorSaveDelay = 5 * time.Second // Actor 存储延迟.
)

// Actor Actor基础封装.
type Actor struct {
	actor.Actor
}

func (a *Actor) GetActor() actor.Actor {
	return a.Actor
}

func (a *Actor) Sugared() Sugared {
	return Sugared{
		Actor: a.Actor,
	}
}

// CActor CActor基础封装.
type CActor struct {
	actor.CActor
}

func (a *CActor) GetActor() actor.Actor {
	return a.CActor
}

func (a *CActor) GetCActor() actor.CActor {
	return a.CActor
}

func (a *CActor) Sugared() CSugared {
	return CSugared{
		CActor: a.CActor,
	}
}

// ActorWithModel 携带数据模型的Actor基础封装.
type ActorWithModel[Model actor.ModelWithDirty] struct {
	Actor
	persistor
	Model Model
}

func NewActorWithModel[Model actor.ModelWithDirty](actor actor.Actor) ActorWithModel[Model] {
	return ActorWithModel[Model]{Actor: Actor{Actor: actor}}
}

func (a *ActorWithModel[Model]) GetModel() actor.Model {
	return a.Model
}

func (a *ActorWithModel[Model]) OnModelDirty() {
	if ok, _ := a.Model.IsDirty(); !ok {
		return
	}
	DelaySave(a, ActorSaveDelay)
}

// CActorWithModel 携带数据模型的CActor基础封装.
type CActorWithModel[Model actor.ModelWithDirty] struct {
	CActor
	persistor
	Model Model
}

func NewCActorWithModel[Model actor.ModelWithDirty](actor actor.CActor) CActorWithModel[Model] {
	return CActorWithModel[Model]{CActor: CActor{CActor: actor}}
}

func (a *CActorWithModel[Model]) GetModel() actor.Model {
	return a.Model
}

func (a *CActorWithModel[Model]) OnModelDirty() {
	if ok, _ := a.Model.IsDirty(); !ok {
		return
	}
	DelaySave(a, ActorSaveDelay)
}

// Sugared Actor 语法糖封装
type Sugared struct {
	actor.Actor
}

func (a Sugared) RPCWithDeadline(to actor.ActorUID, args proto.Message, deadline time.Time) (proto.Message, error) {
	return actorHelper.RPCWithDeadline(a.Actor, to, args, deadline)
}

func (a Sugared) RPCWithTimeout(to actor.ActorUID, args proto.Message, timeout time.Duration) (proto.Message, error) {
	return actorHelper.RPCWithTimeout(a.Actor, to, args, timeout)
}

func (a Sugared) RPC(to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return actorHelper.RPC(a.Actor, to, args)
}

func (a Sugared) RPCWithContext(ctx context.Context, to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return actorHelper.RPCWithContext(ctx, a.Actor, to, args)
}

func (a Sugared) AsyncRPCWithDeadline(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback, deadline time.Time) error {
	return actorHelper.AsyncRPCWithDeadline(a.Actor, to, args, callback, deadline)
}

func (a Sugared) AsyncRPCWithTimeout(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback, timeout time.Duration) error {
	return actorHelper.AsyncRPCWithTimeout(a.Actor, to, args, callback, timeout)
}

func (a Sugared) AsyncRPC(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback) error {
	return actorHelper.AsyncRPC(a.Actor, to, args, callback)
}

func (a Sugared) AsyncRPCWithContext(ctx context.Context, to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback) error {
	return actorHelper.AsyncRPCWithContext(ctx, a.Actor, to, args, callback)
}

// CSugared CActor 语法糖封装
type CSugared struct {
	actor.CActor
}

func (a CSugared) PushRawMessage(msg proto.Message) error {
	return actorHelper.PushRawMessage(a.CActor, msg)
}

func (a CSugared) RPCWithDeadline(to actor.ActorUID, args proto.Message, deadline time.Time) (proto.Message, error) {
	return actorHelper.RPCWithDeadline(a.CActor, to, args, deadline)
}

func (a CSugared) RPCWithTimeout(to actor.ActorUID, args proto.Message, timeout time.Duration) (proto.Message, error) {
	return actorHelper.RPCWithTimeout(a.CActor, to, args, timeout)
}

func (a CSugared) RPC(to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return actorHelper.RPC(a.CActor, to, args)
}

func (a CSugared) RPCWithContext(ctx context.Context, to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return actorHelper.RPCWithContext(ctx, a.CActor, to, args)
}

func (a CSugared) AsyncRPCWithDeadline(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback, deadline time.Time) error {
	return actorHelper.AsyncRPCWithDeadline(a.CActor, to, args, callback, deadline)
}

func (a CSugared) AsyncRPCWithTimeout(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback, timeout time.Duration) error {
	return actorHelper.AsyncRPCWithTimeout(a.CActor, to, args, callback, timeout)
}

func (a CSugared) AsyncRPC(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback) error {
	return actorHelper.AsyncRPC(a.CActor, to, args, callback)
}

func (a CSugared) AsyncRPCWithContext(ctx context.Context, to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback) error {
	return actorHelper.AsyncRPCWithContext(ctx, a.CActor, to, args, callback)
}

// GetModule 通过actor获取模块的通用泛型封装.
func GetModule[M actor.Module](a actor.ActorWithModule, autoCreate bool) M {
	return actor.GetModuleOfActor[M](a, autoCreate)
}
