package actors

import (
	"context"
	"time"

	"github.com/godyy/ggskit/infra/actor"
	"google.golang.org/protobuf/proto"
)

const (
	ctxKeyMsg = "ctx:msg"
)

// CtxSetMsg 设置上下文中的消息.
func CtxSetMsg(ctx *actor.Context, msg proto.Message) {
	ctx.Set(ctxKeyMsg, msg)
}

// CtxGetMsg 获取上下文中的消息.
func CtxGetMsg(ctx *actor.Context) proto.Message {
	if v, ok := ctx.Get(ctxKeyMsg); ok {
		return v.(proto.Message)
	}
	return nil
}

// CtxDecode.
func CtxDecode(ctx *actor.Context) (uint16, proto.Message, error) {
	return contextHelper.Decode(ctx)
}

// CtxReply.
func CtxReply(ctx *actor.Context, msg proto.Message) error {
	return contextHelper.Reply(ctx, msg)
}

// CtxRPCWithDeadline
func CtxRPCWithDeadline(ctx *actor.Context, to actor.ActorUID, args proto.Message, deadline time.Time) (proto.Message, error) {
	return contextHelper.RPCWithDeadline(ctx, to, args, deadline)
}

// CtxRPCWithTimeout
func CtxRPCWithTimeout(ctx *actor.Context, to actor.ActorUID, args proto.Message, timeout time.Duration) (proto.Message, error) {
	return contextHelper.RPCWithTimeout(ctx, to, args, timeout)
}

// CtxRPC
func CtxRPC(ctx *actor.Context, to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return contextHelper.RPC(ctx, to, args)
}

// CtxRPCWithContext
func CtxRPCWithContext(ctx *actor.Context, cctx context.Context, to actor.ActorUID, args proto.Message) (proto.Message, error) {
	return contextHelper.RPCWithContext(ctx, cctx, to, args)
}

// CtxAsyncRPCWithDeadline
func CtxAsyncRPCWithDeadline(ctx *actor.Context, to actor.ActorUID, args proto.Message, callback actor.ContextAsyncRPCCallback, deadline time.Time) error {
	return contextHelper.AsyncRPCWithDeadline(ctx, to, args, callback, deadline)
}

// CtxAsyncRPCWithTimeout
func CtxAsyncRPCWithTimeout(ctx *actor.Context, to actor.ActorUID, args proto.Message, callback actor.ContextAsyncRPCCallback, timeout time.Duration) error {
	return contextHelper.AsyncRPCWithTimeout(ctx, to, args, callback, timeout)
}

// CtxAsyncRPC
func CtxAsyncRPC(ctx *actor.Context, to actor.ActorUID, args proto.Message, callback actor.ContextAsyncRPCCallback) error {
	return contextHelper.AsyncRPC(ctx, to, args, callback)
}

// CtxAsyncRPCWithContext
func CtxAsyncRPCWithContext(ctx *actor.Context, cctx context.Context, to actor.ActorUID, args proto.Message, callback actor.ContextAsyncRPCCallback) error {
	return contextHelper.AsyncRPCWithContext(ctx, cctx, to, args, callback)
}
