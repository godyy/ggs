package handlers

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/actor"
	"google.golang.org/protobuf/proto"
)

const (
	paramKeyPayload = "handlers:payload"
)

// GetActor 获取当前上下文的Actor.
func GetActor[Actor gactor.ActorBehavior](ctx *gactor.Context) Actor {
	return ctx.Actor().Behavior().(Actor)
}

// setC2SPayload 设置当前上下文的C2SPayload.
func setC2SPayload(ctx *gactor.Context, payload actor.C2SPayload) {
	ctx.Set(paramKeyPayload, payload)
}

// getC2SPayload 获取当前上下文的C2SPayload.
func getC2SPayload(ctx *gactor.Context) (actor.C2SPayload, bool) {
	if value, exists := ctx.Get(paramKeyPayload); exists {
		return value.(actor.C2SPayload), true
	} else {
		return actor.C2SPayload{}, false
	}
}

// setS2SPayload 设置当前上下文的S2SPayload.
func setS2SPayload(ctx *gactor.Context, payload actor.S2SPayload) {
	ctx.Set(paramKeyPayload, payload)
}

// getS2SPayload 获取当前上下文的S2SPayload.
func getS2SPayload(ctx *gactor.Context) (actor.S2SPayload, bool) {
	if value, exists := ctx.Get(paramKeyPayload); exists {
		return value.(actor.S2SPayload), true
	} else {
		return actor.S2SPayload{}, false
	}
}

// getC2SReq 获取当前上下文的C2S请求.
func getC2SReq[Req proto.Message](ctx *gactor.Context) Req {
	payload, _ := getC2SPayload(ctx)
	return payload.Msg.(Req)
}

// getS2SReq 获取当前上下文的S2S请求.
func getS2SReq[Req proto.Message](ctx *gactor.Context) Req {
	payload, _ := getS2SPayload(ctx)
	return payload.Msg.(Req)
}
