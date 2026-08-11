package handler

import (
	"github.com/godyy/ggs/internal/infra/actor"
	pbcommon "github.com/godyy/ggs/internal/infra/actor/protocol/pb/common"
	"google.golang.org/protobuf/proto"
)

// WrapReqFunc 包装Req处理函数.
func WrapReqFunc[Req, Resp proto.Message](f func(ctx *actor.Context, req Req) (Resp, error)) actor.HandlerFunc {
	return func(ctx *actor.Context) {
		defer handlePushMsgQueue(ctx)
		req := GetArgs[Req](ctx)
		resp, err := f(ctx, req)
		if err != nil {
			ReplyErrorAbort(ctx, err)
			return
		}
		Reply(ctx, resp)
	}
}

// WrapReqNoRespFunc 包装Req处理函数, 无返回值.
func WrapReqNoRespFunc[Req proto.Message](f func(ctx *actor.Context, req Req) error) actor.HandlerFunc {
	return WrapReqFunc(func(ctx *actor.Context, req Req) (*pbcommon.Success, error) {
		if err := f(ctx, req); err != nil {
			return nil, err
		}
		return &pbcommon.Success{}, nil
	})
}

// WrapAsyncReqFunc 包装异步Req处理函数.
func WrapAsyncReqFunc[Req proto.Message](f func(ctx *actor.Context, req Req) error) actor.HandlerFunc {
	return func(ctx *actor.Context) {
		req := GetArgs[Req](ctx)
		if err := f(ctx, req); err != nil {
			ReplyErrorAbort(ctx, err)
			return
		}
	}
}

// WrapRPCFunc 包装RPC处理函数.
func WrapRPCFunc[Req, Resp proto.Message](f func(ctx *actor.Context, req Req) (Resp, error)) actor.HandlerFunc {
	return func(ctx *actor.Context) {
		req := GetArgs[Req](ctx)
		resp, err := f(ctx, req)
		if err != nil {
			ReplyErrorAbort(ctx, err)
			return
		}
		Reply(ctx, resp)
	}
}

// WrapRPCNoRespFunc 包装RPC处理函数, 无返回值.
func WrapRPCNoRespFunc[Req proto.Message](f func(ctx *actor.Context, req Req) error) actor.HandlerFunc {
	return WrapRPCFunc(func(ctx *actor.Context, req Req) (*pbcommon.Success, error) {
		if err := f(ctx, req); err != nil {
			return nil, err
		}
		return &pbcommon.Success{}, nil
	})
}

// WrapAsyncRPCFunc 包装异步RPC处理函数.
func WrapAsyncRPCFunc[Req proto.Message](f func(ctx *actor.Context, req Req) error) actor.HandlerFunc {
	return func(ctx *actor.Context) {
		req := GetArgs[Req](ctx)
		if err := f(ctx, req); err != nil {
			ReplyErrorAbort(ctx, err)
			return
		}
	}
}

// WrapCastFunc 包装Cast处理函数.
func WrapCastFunc[Params proto.Message](f func(ctx *actor.Context, params Params) bool) actor.HandlerFunc {
	return func(ctx *actor.Context) {
		params := GetArgs[Params](ctx)
		if !f(ctx, params) {
			ctx.Abort()
		}
	}
}

// Reply 回复消息.
func Reply(ctx *actor.Context, msg proto.Message) {
	actor.SugarContext(ctx).Reply(msg)
}

// ReplySuccess 回复成功.
func ReplySuccess(ctx *actor.Context) {
	Reply(ctx, &pbcommon.Success{})
}

// ReplyError 回复错误.
func ReplyError(ctx *actor.Context, err error) {
	var respErr *pbcommon.Error

	switch e := err.(type) {
	case *actor.PbError:
		respErr = e.Err
	default:
		loggerInst.Errorf("ReplyError: none PbError, %v", err)
		respErr = &pbcommon.Error{Code: pbcommon.ErrCode_ECInternalError}
	}

	actor.SugarContext(ctx).Reply(respErr)
}

// ReplyErrorAbort 回复错误并Abort.
func ReplyErrorAbort(ctx *actor.Context, err error) {
	ReplyError(ctx, err)
	ctx.Abort()
}

// handlePushMsgQueue 处理推送消息队列.
func handlePushMsgQueue(ctx *actor.Context) {
	sugared := actor.CSugared{CActor: ctx.Actor().(actor.CActor)}
	msgQueue, _ := actor.CtxKGet(ctx, ctxKeyPushMsgQueue)
	for _, msg := range msgQueue {
		sugared.PushRawMessage(msg)
	}
}
