package handlers

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/codec"
	_ "github.com/godyy/ggs/app/game/internal/systems"
	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
	pbs2s "github.com/godyy/ggs/internal/proto/pb/s2s"
)

// Handler 消息处理器基础实现.
type Handler struct {
	funcs map[uint16]gactor.HandlerFunc // 协议ID到处理函数的映射
}

// NewHandler 创建一个新的 Handler.
func NewHandler() *Handler {
	return &Handler{
		funcs: make(map[uint16]gactor.HandlerFunc),
	}
}

// RegisterFunc 注册消息处理函数.
func (h *Handler) RegisterFunc(pid uint16, funcs ...gactor.HandlerFunc) bool {
	if _, ok := h.funcs[pid]; ok {
		return false
	}
	h.funcs[pid] = gactor.NewHandlersChain(funcs...).Handle
	return true
}

// getFunc 获取PID对应的处理函数.
func (h *Handler) getFunc(pid uint16) gactor.HandlerFunc {
	return h.funcs[pid]
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
	// 解码负载数据
	var payload codec.C2SPayload
	if err := ctx.Decode(&payload); err != nil {
		ctx.ReplyDecodeError()
		ctx.Abort()
		return
	}

	// todo 熔断逻辑

	// 获取处理函数
	f := h.getFunc(payload.PID)
	if f == nil {
		// todo
		ctx.Abort()
		return
	}

	// 调用处理函数
	setC2SPayload(ctx, payload)
	f(ctx)
}

// handleS2S 处理S2S请求.
func (h *Handler) handleS2S(ctx *gactor.Context) {
	// 解码负载数据
	var payload codec.S2SPayload
	if err := ctx.Decode(&payload); err != nil {
		ctx.ReplyDecodeError()
		ctx.Abort()
		return
	}

	// 获取处理函数
	f := h.getFunc(payload.PID)
	if f == nil {
		// todo
		ctx.Abort()
		return
	}

	// 调用处理函数
	setS2SPayload(ctx, payload)
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
