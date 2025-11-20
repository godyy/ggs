package handlers

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/codec"
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
func setC2SPayload(ctx *gactor.Context, payload codec.C2SPayload) {
	ctx.Set(paramKeyPayload, payload)
}

// getC2SPayload 获取当前上下文的C2SPayload.
func getC2SPayload(ctx *gactor.Context) (codec.C2SPayload, bool) {
	if value, exists := ctx.Get(paramKeyPayload); exists {
		return value.(codec.C2SPayload), true
	} else {
		return codec.C2SPayload{}, false
	}
}

// setS2SPayload 设置当前上下文的S2SPayload.
func setS2SPayload(ctx *gactor.Context, payload codec.S2SPayload) {
	ctx.Set(paramKeyPayload, payload)
}

// getS2SPayload 获取当前上下文的S2SPayload.
func getS2SPayload(ctx *gactor.Context) (codec.S2SPayload, bool) {
	if value, exists := ctx.Get(paramKeyPayload); exists {
		return value.(codec.S2SPayload), true
	} else {
		return codec.S2SPayload{}, false
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
