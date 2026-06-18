package handler

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actor/handler"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	pbs2s "github.com/godyy/ggs/internal/protocol/pb/s2s"
	"github.com/godyy/ggskit/infra/actor"
)

var (
	c2sHandler = handler.NewC2SHandler()
	s2sHandler = handler.NewS2SHandler()
)

// Hanlde 请求处理函数入口
func Handle(ctx *actor.Context) {
	if ctx.RequestType() == gactor.RequestTypeReq {
		c2sHandler.Handle(ctx)
	} else {
		s2sHandler.Handle(ctx)
	}
}

// RegisterC2S 注册C2S请求处理函数
func RegisterC2S(pid pbc2s.PID, funcs ...handler.HandlerFunc) {
	c2sHandler.RegisterFunc(pid, funcs...)
}

// RegisterS2S 注册S2S请求函数
func RegisterS2S(pid pbs2s.PID, funcs ...handler.HandlerFunc) {
	s2sHandler.RegisterFunc(pid, funcs...)
}
