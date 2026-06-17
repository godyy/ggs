package actor

import (
	"context"
	"time"

	"github.com/godyy/ggskit/infra/actor"
	"google.golang.org/protobuf/proto"
)

type Context = actor.Context

func CtxActor[Actor actor.ActorBehavior](ctx *Context) Actor {
	return actor.CtxActor[Actor](ctx)
}

// contextSugarUtil 全局上下文语法糖工具.
var contextSugarUtil *actor.ContextHelper

const (
	ctxKeyMsg = "ctx:msg"
)

// ContextSugared 上下文语法糖封装.
type ContextSugared struct {
	*Context
}

// SugarContext 给上下文加糖.
func SugarContext(ctx *Context) ContextSugared {
	return ContextSugared{Context: ctx}
}

// SetMsg 设置上下文中的消息.
func (ctx ContextSugared) SetMsg(msg proto.Message) {
	ctx.Set(ctxKeyMsg, msg)
}

// GetMsg 获取上下文中的消息.
func (ctx ContextSugared) GetMsg() proto.Message {
	if v, ok := ctx.Get(ctxKeyMsg); ok {
		return v.(proto.Message)
	}
	return nil
}

// Decode
func (ctx ContextSugared) Decode() (uint16, proto.Message, error) {
	return contextSugarUtil.Decode(ctx.Context)
}

// Reply
func (ctx ContextSugared) Reply(msg proto.Message) error {
	return contextSugarUtil.Reply(ctx.Context, msg)
}

// RPCWithDeadline
func (ctx ContextSugared) RPCWithDeadline(to ActorUID, args proto.Message, deadline time.Time) (proto.Message, error) {
	return contextSugarUtil.RPCWithDeadline(ctx.Context, to, args, deadline)
}

// RPCWithTimeout
func (ctx ContextSugared) RPCWithTimeout(to ActorUID, args proto.Message, timeout time.Duration) (proto.Message, error) {
	return contextSugarUtil.RPCWithTimeout(ctx.Context, to, args, timeout)
}

// RPC
func (ctx ContextSugared) RPC(to ActorUID, args proto.Message) (proto.Message, error) {
	return contextSugarUtil.RPC(ctx.Context, to, args)
}

// RPCWithContext
func (ctx ContextSugared) RPCWithContext(cctx context.Context, to ActorUID, args proto.Message) (proto.Message, error) {
	return contextSugarUtil.RPCWithContext(ctx.Context, cctx, to, args)
}

// AsyncRPCWithDeadline
func (ctx ContextSugared) AsyncRPCWithDeadline(to ActorUID, args proto.Message, callback actor.ContextAsyncRPCCallback, deadline time.Time) error {
	return contextSugarUtil.AsyncRPCWithDeadline(ctx.Context, to, args, callback, deadline)
}

// AsyncRPCWithTimeout
func (ctx ContextSugared) AsyncRPCWithTimeout(to ActorUID, args proto.Message, callback actor.ContextAsyncRPCCallback, timeout time.Duration) error {
	return contextSugarUtil.AsyncRPCWithTimeout(ctx.Context, to, args, callback, timeout)
}

// AsyncRPC
func (ctx ContextSugared) AsyncRPC(to ActorUID, args proto.Message, callback actor.ContextAsyncRPCCallback) error {
	return contextSugarUtil.AsyncRPC(ctx.Context, to, args, callback)
}

// AsyncRPCWithContext
func (ctx ContextSugared) AsyncRPCWithContext(cctx context.Context, to ActorUID, args proto.Message, callback actor.ContextAsyncRPCCallback) error {
	return contextSugarUtil.AsyncRPCWithContext(ctx.Context, cctx, to, args, callback)
}

// Cast
func (ctx ContextSugared) Cast(to ActorUID, msg proto.Message) error {
	return contextSugarUtil.Cast(ctx.Context, to, msg)
}
