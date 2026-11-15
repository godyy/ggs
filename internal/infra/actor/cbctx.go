package actor

import (
	"github.com/godyy/gactor"
	"github.com/godyy/gutils/container/kv"
)

// CbCtx 回调上下文映射.
type CbCtx = kv.M

// CbCtxK 回调上下文键.
type CbCtxK[V any] = kv.K[V]

// NewCbCtx 创建回调上下文映射.
func NewCbCtx(pairs ...any) *CbCtx {
	return gactor.NewCbCtx(pairs...)
}

// CbCtxRelease 释放回调上下文映射.
func CbCtxRelease(ctx *CbCtx) {
	gactor.CbCtxRelease(ctx)
}

// NewCbCtxK 创建回调上下文键.
func NewCbCtxK[V any](s string) CbCtxK[V] {
	return kv.NewK[V](s)
}

// CbCtxGet 获取回调上下文映射中的值.
func CbCtxGet[V any](ctx *CbCtx, k CbCtxK[V]) (V, bool) {
	return kv.Get(ctx, k)
}

// CbCtxMust 获取回调上下文映射中的值,若值不存在则panic.
func CbCtxMust[V any](ctx *CbCtx, k CbCtxK[V]) V {
	return kv.Must(ctx, k)
}
