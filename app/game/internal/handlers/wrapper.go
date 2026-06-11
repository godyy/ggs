package handlers

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggskit/infra/actor"

	"github.com/godyy/ggs/app/game/internal/base/errs"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/protocol/pb/common"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
	protoreg "github.com/godyy/ggs/internal/protocol/registry"
	codecc2s "github.com/godyy/ggskit/base/codec/c2s"
	"google.golang.org/protobuf/proto"
)

// WrapC2SFunc 包装C2S函数.
func WrapC2SFunc[Req, Resp proto.Message](f func(ctx *gactor.Context, req Req) (Resp, error)) gactor.HandlerFunc {
	return func(ctx *gactor.Context) {
		req := getC2SReq[Req](ctx)
		resp, err := f(ctx, req)
		if err != nil {
			replyC2SError(ctx, err)
			return
		}
		replyC2S(ctx, resp)
	}
}

// WrapS2SRPCFunc 包装S2S RPC函数.
func WrapS2SRPCFunc[Req, Resp proto.Message](f func(ctx *gactor.Context, req Req) (Resp, error)) gactor.HandlerFunc {
	return func(ctx *gactor.Context) {
		req := getS2SReq[Req](ctx)
		resp, err := f(ctx, req)
		if err != nil {
			replyS2SError(ctx, err)
			return
		}
		replyS2S(ctx, resp)
	}
}

// WrapS2SCastFunc 包装S2S Cast函数.
func WrapS2SCastFunc[Params proto.Message](f func(ctx *gactor.Context, params Params)) gactor.HandlerFunc {
	return func(ctx *gactor.Context) {
		params := getS2SReq[Params](ctx)
		f(ctx, params)
	}
}

// replyC2S 回复C2S请求.
func replyC2S(ctx *gactor.Context, resp proto.Message) {
	reqPayload, _ := getC2SPayload(ctx)
	respPayload := actor.C2SPayload{
		Pt:  codecc2s.PtResp,
		Seq: reqPayload.Seq,
	}
	pid, ok := protoreg.C2S.GetPid(resp)
	if !ok {
		loggerInst.Errorf("replyC2S: pid not found for resp %T", resp)
		return
	}
	respPayload.PID = pid
	respPayload.Msg = resp
	if err := ctx.Reply(&respPayload); err != nil {
		loggerInst.Errorf("replyC2S: failed to reply %T, err: %v", resp, err)
		return
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
	replyC2S(ctx, respErr)
}

// replyS2S 回复S2S请求.
func replyS2S(ctx *gactor.Context, resp proto.Message) {
	respPayload := actor.S2SPayload{}
	pid, ok := protoreg.S2S.GetPid(resp)
	if !ok {
		loggerInst.Errorf("replyS2S: pid not found for resp %T", resp)
		return
	}
	respPayload.PID = pid
	respPayload.Msg = resp
	if err := ctx.Reply(&respPayload); err != nil {
		loggerInst.Errorf("replyC2S: failed to reply %T, err: %v", resp, err)
		return
	}
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
	replyC2S(ctx, respErr)
}
