package actors

import (
	"context"
	"time"

	"github.com/godyy/ggskit/infra/actor"
	"google.golang.org/protobuf/proto"
)

// actorSugarUtil Actor 语法糖工具.
var actorSugarUtil *actor.ActorHelper

// ActorSugared Actor和语法糖基础封装.
type ActorSugared struct {
	actor.Actor
}

func (a *ActorSugared) GetActor() actor.Actor {
	return a.Actor
}

func (a *ActorSugared) Sugared() Sugared {
	return Sugared{
		Actor: a.Actor,
	}
}

// CActorSugared CActor和语法糖基础封装.
type CActorSugared struct {
	actor.CActor
}

func (a *CActorSugared) GetActor() actor.Actor {
	return a.CActor
}

func (a *CActorSugared) GetCActor() actor.CActor {
	return a.CActor
}

func (a *CActorSugared) Sugared() CSugared {
	return CSugared{
		CActor: a.CActor,
	}
}

// Sugared Actor 语法糖封装
type Sugared struct {
	actor.Actor
}

func (a Sugared) RPCWithDeadline(to actor.ActorUID, args proto.Message, deadline time.Time) (proto.Message, error) {
	return actorSugarUtil.RPCWithDeadline(a.Actor, to, args, deadline)
}

func (a Sugared) RPCWithTimeout(to actor.ActorUID, args proto.Message, timeout time.Duration) (proto.Message, error) {
	return actorSugarUtil.RPCWithTimeout(a.Actor, to, args, timeout)
}

func (a Sugared) RPC(to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return actorSugarUtil.RPC(a.Actor, to, args)
}

func (a Sugared) RPCWithContext(ctx context.Context, to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return actorSugarUtil.RPCWithContext(ctx, a.Actor, to, args)
}

func (a Sugared) AsyncRPCWithDeadline(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback, deadline time.Time) error {
	return actorSugarUtil.AsyncRPCWithDeadline(a.Actor, to, args, callback, deadline)
}

func (a Sugared) AsyncRPCWithTimeout(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback, timeout time.Duration) error {
	return actorSugarUtil.AsyncRPCWithTimeout(a.Actor, to, args, callback, timeout)
}

func (a Sugared) AsyncRPC(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback) error {
	return actorSugarUtil.AsyncRPC(a.Actor, to, args, callback)
}

func (a Sugared) AsyncRPCWithContext(ctx context.Context, to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback) error {
	return actorSugarUtil.AsyncRPCWithContext(ctx, a.Actor, to, args, callback)
}

// CSugared CActor 语法糖封装
type CSugared struct {
	actor.CActor
}

func (a CSugared) PushRawMessage(msg proto.Message) error {
	return actorSugarUtil.PushRawMessage(a.CActor, msg)
}

func (a CSugared) RPCWithDeadline(to actor.ActorUID, args proto.Message, deadline time.Time) (proto.Message, error) {
	return actorSugarUtil.RPCWithDeadline(a.CActor, to, args, deadline)
}

func (a CSugared) RPCWithTimeout(to actor.ActorUID, args proto.Message, timeout time.Duration) (proto.Message, error) {
	return actorSugarUtil.RPCWithTimeout(a.CActor, to, args, timeout)
}

func (a CSugared) RPC(to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return actorSugarUtil.RPC(a.CActor, to, args)
}

func (a CSugared) RPCWithContext(ctx context.Context, to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return actorSugarUtil.RPCWithContext(ctx, a.CActor, to, args)
}

func (a CSugared) AsyncRPCWithDeadline(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback, deadline time.Time) error {
	return actorSugarUtil.AsyncRPCWithDeadline(a.CActor, to, args, callback, deadline)
}

func (a CSugared) AsyncRPCWithTimeout(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback, timeout time.Duration) error {
	return actorSugarUtil.AsyncRPCWithTimeout(a.CActor, to, args, callback, timeout)
}

func (a CSugared) AsyncRPC(to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback) error {
	return actorSugarUtil.AsyncRPC(a.CActor, to, args, callback)
}

func (a CSugared) AsyncRPCWithContext(ctx context.Context, to actor.ActorUID, args proto.Message, callback actor.ActorAsyncRPCCallback) error {
	return actorSugarUtil.AsyncRPCWithContext(ctx, a.CActor, to, args, callback)
}
