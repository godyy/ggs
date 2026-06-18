package handler

import (
	"github.com/godyy/ggs/internal/infra/actor"
	"google.golang.org/protobuf/proto"
)

var (
	ctxKeyPushMsgQueue = actor.NewCtxK[[]proto.Message]() // 推送的消息队列
)

// GetArgs 获取当前上下文的请求参数.
func GetArgs[Args proto.Message](ctx *actor.Context) Args {
	return actor.SugarContext(ctx).GetMsg().(Args)
}

// AppendPushMsg 追加推送消息.
func AppendPushMsg(ctx *actor.Context, msg proto.Message) {
	msgQueue, _ := actor.CtxKGet(ctx, ctxKeyPushMsgQueue)
	msgQueue = append(msgQueue, msg)
	actor.CtxKSet(ctx, ctxKeyPushMsgQueue, msgQueue)
}
