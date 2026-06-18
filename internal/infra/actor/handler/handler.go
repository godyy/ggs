package handler

import (
	iactor "github.com/godyy/ggs/internal/infra/actor"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/protocol/pb/common"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
	"github.com/godyy/ggskit/infra/actor"
)

type HandlerFunc = actor.HandlerFunc

// C2SHandler C2S消息处理器.
type C2SHandler struct {
	m actor.HandlerMap
}

func NewC2SHandler() *C2SHandler {
	return &C2SHandler{
		m: make(actor.HandlerMap),
	}
}

func (h *C2SHandler) RegisterFunc(pid pbc2s.PID, funcs ...HandlerFunc) bool {
	return h.m.Register(uint16(pid), funcs...)
}

func (h *C2SHandler) Handle(ctx *actor.Context) {
	ctxSugared := iactor.SugarContext(ctx)

	// 解码负载数据
	pid, msg, err := ctxSugared.Decode()
	if err != nil {
		ctx.ReplyDecodeError()
		ctx.Abort()
		return
	}
	ctxSugared.SetMsg(msg)

	// todo 熔断逻辑

	// 获取处理函数
	f := h.m[pid]
	if f == nil {
		ctxSugared.Reply(&pbcommon.Error{
			Code: int32(pbc2s.ErrCode_ECInternalError),
		})
		ctx.Abort()
		return
	}

	// 调用处理函数
	f(ctx)
}

// S2SHandler S2S消息处理器.
type S2SHandler struct {
	m actor.HandlerMap
}

func NewS2SHandler() *S2SHandler {
	return &S2SHandler{
		m: make(actor.HandlerMap),
	}
}

func (h *S2SHandler) RegisterFunc(pid pbs2s.PID, funcs ...HandlerFunc) bool {
	return h.m.Register(uint16(pid), funcs...)
}

func (h *S2SHandler) Handle(ctx *actor.Context) {
	ctxSugared := iactor.SugarContext(ctx)

	// 解码负载数据
	pid, msg, err := ctxSugared.Decode()
	if err != nil {
		ctx.ReplyDecodeError()
		ctx.Abort()
		return
	}
	ctxSugared.SetMsg(msg)

	// todo 熔断逻辑

	// 获取处理函数
	f := h.m[pid]
	if f == nil {
		ctxSugared.Reply(&pbcommon.Error{
			Code: int32(pbs2s.ErrCode_ECInternalError),
		})
		ctx.Abort()
		return
	}

	// 调用处理函数
	f(ctx)
}
