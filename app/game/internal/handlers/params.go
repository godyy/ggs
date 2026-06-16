package handlers

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actors"
	"google.golang.org/protobuf/proto"
)

const (
	paramKeyPayload = "handlers:payload"
)

// getReq 获取当前上下文的请求消息.
func getReq[Req proto.Message](ctx *gactor.Context) Req {
	return actors.CtxGetMsg(ctx).(Req)
}
