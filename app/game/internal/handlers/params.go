package handlers

import (
	"github.com/godyy/ggs/internal/infra/actor"
	"google.golang.org/protobuf/proto"
)

const (
	paramKeyPayload = "handlers:payload"
)

// getReq 获取当前上下文的请求消息.
func getReq[Req proto.Message](ctx *actor.Context) Req {
	return actor.SugarContext(ctx).GetMsg().(Req)
}
