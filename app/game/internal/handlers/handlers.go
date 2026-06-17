package handlers

import (
	"github.com/godyy/gactor"
	_ "github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/infra/actors"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
	"github.com/godyy/ggskit/infra/actor"
)

// Handler 消息处理器基础实现.
type Handler struct {
	hm actor.HandlerMap
}

// NewHandler 创建一个新的 Handler.
func NewHandler() *Handler {
	return &Handler{
		hm: make(actor.HandlerMap),
	}
}

// RegisterFunc 注册消息处理函数.
func (h *Handler) RegisterFunc(pid uint16, funcs ...gactor.HandlerFunc) bool {
	return h.hm.Register(pid, funcs...)
}

// getFunc 获取PID对应的处理函数.
func (h *Handler) getFunc(pid uint16) gactor.HandlerFunc {
	return h.hm[pid]
}

// Handle 处理请求.
func (h *Handler) Handle(ctx *gactor.Context) {
	if ctx.RequestType() == gactor.RequestTypeReq {
		h.handleC2S(ctx)
	} else {
		h.handleS2S(ctx)
	}
}

// handleC2S 处理C2S请求.
func (h *Handler) handleC2S(ctx *gactor.Context) {
	ctxSugared := actors.SugarContext(ctx)

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
	f := h.getFunc(pid)
	if f == nil {
		// todo
		ctx.Abort()
		return
	}

	// 调用处理函数
	f(ctx)
}

// handleS2S 处理S2S请求.
func (h *Handler) handleS2S(ctx *gactor.Context) {
	ctxSugared := actors.SugarContext(ctx)

	// 解码负载数据
	pid, msg, err := ctxSugared.Decode()
	if err != nil {
		ctx.ReplyDecodeError()
		ctx.Abort()
		return
	}
	ctxSugared.SetMsg(msg)

	// 获取处理函数
	f := h.getFunc(pid)
	if f == nil {
		// todo
		ctx.Abort()
		return
	}

	// 调用处理函数
	f(ctx)
}

// RegisterC2SFunc 注册C2S消息处理函数.
func RegisterC2SFunc(h *Handler, pid pbc2s.PID, funcs ...gactor.HandlerFunc) bool {
	return h.RegisterFunc(uint16(pid), funcs...)
}

// RegisterS2SFunc 注册S2S消息处理函数.
func RegisterS2SFunc(h *Handler, pid pbs2s.PID, funcs ...gactor.HandlerFunc) bool {
	return h.RegisterFunc(uint16(pid), funcs...)
}
