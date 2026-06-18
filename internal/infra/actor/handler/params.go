package handler

import (
	"github.com/godyy/ggs/internal/infra/actor"
	"google.golang.org/protobuf/proto"
)

// GetArgs 获取当前上下文的请求参数.
func GetArgs[Args proto.Message](ctx *actor.Context) Args {
	return actor.SugarContext(ctx).GetMsg().(Args)
}
