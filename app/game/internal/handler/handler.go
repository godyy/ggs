package handler

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actor/handler"
	"github.com/godyy/ggskit/infra/actor"
	"google.golang.org/protobuf/proto"
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
func RegisterC2S(msg proto.Message, funcs ...handler.HandlerFunc) {
	c2sHandler.RegisterFunc(msg, funcs...)
}

// RegisterS2S 注册S2S请求函数
func RegisterS2S(msg proto.Message, funcs ...handler.HandlerFunc) {
	s2sHandler.RegisterFunc(msg, funcs...)
}
