package handler

import (
	"github.com/godyy/ggs/internal/infra/actor"
	pbc2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/infra/actor/protocol/pb/common"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
	"google.golang.org/protobuf/proto"
)

// WrapReqFunc 包装Req处理函数.
func WrapReqFunc[Req, Resp proto.Message](f func(ctx *actor.Context, req Req) (Resp, error)) HandlerFunc {
	return func(ctx *actor.Context) {
		req := GetArgs[Req](ctx)
		resp, err := f(ctx, req)
		if err != nil {
			replyC2SError(ctx, err)
			ctx.Abort()
			return
		}
		actor.SugarContext(ctx).Reply(resp)
	}
}

// WrapRPCFunc 包装RPC处理函数.
func WrapRPCFunc[Req, Resp proto.Message](f func(ctx *actor.Context, req Req) (Resp, error)) HandlerFunc {
	return func(ctx *actor.Context) {
		req := GetArgs[Req](ctx)
		resp, err := f(ctx, req)
		if err != nil {
			replyS2SError(ctx, err)
			ctx.Abort()
			return
		}
		actor.SugarContext(ctx).Reply(resp)
	}
}

// WrapCastFunc 包装Cast处理函数.
func WrapCastFunc[Params proto.Message](f func(ctx *actor.Context, params Params) bool) HandlerFunc {
	return func(ctx *actor.Context) {
		params := GetArgs[Params](ctx)
		if !f(ctx, params) {
			ctx.Abort()
		}
	}
}

// replyC2SError 回复C2S错误.
func replyC2SError(ctx *actor.Context, err error) {
	var respErr *pbcommon.Error

	switch e := err.(type) {
	case *PbError:
		respErr = e.Err
	default:
		loggerInst.Errorf("replyC2SError: none PbError, %v", err)
		respErr = &pbcommon.Error{Code: int32(pbc2s.ErrCode_ECInternalError)}
	}

	actor.SugarContext(ctx).Reply(respErr)
}

// replyS2SError 回复S2S错误.
func replyS2SError(ctx *actor.Context, err error) {
	var respErr *pbcommon.Error
	switch e := err.(type) {
	case *PbError:
		respErr = e.Err
	default:
		loggerInst.Errorf("replyS2SError: none PbError, %v", err)
		respErr = &pbcommon.Error{Code: int32(pbs2s.ErrCode_ECInternalError)}
	}
	actor.SugarContext(ctx).Reply(respErr)
}
