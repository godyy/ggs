package handlers

import (
	"github.com/godyy/gactor"

	"github.com/godyy/ggs/app/game/internal/base/errs"
	"github.com/godyy/ggs/internal/infra/actor"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/protocol/pb/common"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
	"google.golang.org/protobuf/proto"
)

// WrapC2SFunc 包装C2S函数.
func WrapC2SFunc[Req, Resp proto.Message](f func(ctx *gactor.Context, req Req) (Resp, error)) gactor.HandlerFunc {
	return func(ctx *gactor.Context) {
		req := getReq[Req](ctx)
		resp, err := f(ctx, req)
		if err != nil {
			replyC2SError(ctx, err)
			return
		}
		actor.SugarContext(ctx).Reply(resp)
	}
}

// WrapS2SRPCFunc 包装S2S RPC函数.
func WrapS2SRPCFunc[Req, Resp proto.Message](f func(ctx *gactor.Context, req Req) (Resp, error)) gactor.HandlerFunc {
	return func(ctx *gactor.Context) {
		req := getReq[Req](ctx)
		resp, err := f(ctx, req)
		if err != nil {
			replyS2SError(ctx, err)
			return
		}
		actor.SugarContext(ctx).Reply(resp)
	}
}

// WrapS2SCastFunc 包装S2S Cast函数.
func WrapS2SCastFunc[Params proto.Message](f func(ctx *gactor.Context, params Params)) gactor.HandlerFunc {
	return func(ctx *gactor.Context) {
		params := getReq[Params](ctx)
		f(ctx, params)
	}
}

// replyC2SError 回复C2S错误.
func replyC2SError(ctx *gactor.Context, err error) {
	var respErr *pbcommon.Error

	switch e := err.(type) {
	case *errs.PbError:
		respErr = e.Err
	default:
		loggerInst.Errorf("replyC2SError: none PbError, %v", err)
		respErr = &pbcommon.Error{Code: int32(pbc2s.ErrCode_ECInternalError)}
	}

	actor.SugarContext(ctx).Reply(respErr)
}

// replyS2SError 回复S2S错误.
func replyS2SError(ctx *gactor.Context, err error) {
	var respErr *pbcommon.Error
	switch e := err.(type) {
	case *errs.PbError:
		respErr = e.Err
	default:
		loggerInst.Errorf("replyS2SError: none PbError, %v", err)
		respErr = &pbcommon.Error{Code: int32(pbs2s.ErrCode_ECInternalError)}
	}
	actor.SugarContext(ctx).Reply(respErr)
}
