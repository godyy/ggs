---
name: ggs-actor-handler
description: Create and wire ggs actor message handlers including Wrap func selection, reply/error/push/logger helpers, and CtxK context keys. Use when adding or registering c2s/s2s handlers for an actor. Not for proto or systems code.
---

# GGS Actor 消息处理器 Playbook

把“新增/修改一个 Actor 消息处理器”标准化。核心原则：**handler 只做协议适配**——pb 参数校验、调用 systems、把领域结果映射回 pb；不写业务规则、不碰模块状态。

本技能覆盖三件事：**选哪种类别（Wrap 函数）**、**能用哪些辅助功能**、**如何用 CtxK 存取处理过程中的上下文参数**。协议定义见 `ggs-add-protocol`，跨异步回调传参见 `ggs-cbctx`，端到端功能新增见 `ggs-actor-feature`。

## 0. 落点与注册入口

```text
internal/infra/actor/handler/     # 通用 wrapper 与辅助函数（根级共享层）
├── wrapper.go                    # Wrap* 包装器 + Reply* 辅助
├── params.go                     # GetArgs / AppendPushMsg / push 队列 key
├── handler.go                    # C2SHandler / S2SHandler，按 pid 分发
└── log.go                        # handler 专用 logger

app/<service>/internal/handler/   # 服务内 handler 实现
├── handler.go                    # RegisterC2S / RegisterS2S 入口 + Handle 分发
└── <actor>/                      # 按 Actor 分目录：player/ server/ chat/ common/
    ├── handler.go                # init() 内登记消息，是唯一注册点
    └── <domain>.go               # 按功能域拆处理函数
```

- **按消息类型注册，不按 PID**：`RegisterC2S(msg, funcs...)` / `RegisterS2S(msg, funcs...)`，PID 由消息全名哈希推导（见 `ggs-add-protocol`）。
- `Handle(ctx)` 按 `ctx.RequestType()` 决定走 c2s 还是 s2s 分发；**注册到错误的通道 = 运行时找不到 handler → `ECInternalError`**。
- 每个 Actor 一个目录，每个功能域一个文件，避免单文件膨胀。
- 注册点统一写在 `<actor>/handler.go` 的 `init()`，便于被 `import _` 触发。

参考：[handler.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/handler/handler.go)、[player/handler.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/handler/player/handler.go)。

## 1. Handler 类别（选哪个 Wrap 函数）

wrapper 定义在 [wrapper.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/handler/wrapper.go)。**选型只看两个维度：方向（c2s / s2s）和同步方式（同步 / 异步 / 单向）**。

| 场景 | 协议 | wrapper | 处理函数签名 | 自动 Reply |
| --- | --- | --- | --- | --- |
| c2s 同步请求 | c2s `Req`（有 `Resp`） | `WrapReqFunc` | `func(ctx, req Req) (Resp, error)` | 是 |
| c2s 同步请求（无返回体） | c2s `Req` | `WrapReqNoRespFunc` | `func(ctx, req Req) error` | 是（回 `Success`） |
| c2s 异步请求 | c2s `Req` | `WrapAsyncReqFunc` | `func(ctx, req Req) error` | 否，回调里自己 Reply |
| s2s 同步 RPC | s2s `Req`（有 `Resp`） | `WrapRPCFunc` | `func(ctx, req Req) (Resp, error)` | 是 |
| s2s 同步 RPC（无返回体） | s2s `Req` | `WrapRPCNoRespFunc` | `func(ctx, req Req) error` | 是（回 `Success`） |
| s2s 异步 RPC | s2s `Req` | `WrapAsyncRPCFunc` | `func(ctx, req Req) error` | 否，回调里自己 Reply |
| s2s 单向通知 / Cast | s2s `Ntf` | `WrapCastFunc` | `func(ctx, params Params) bool` | 不适用 |

**必须记住的差异（选错会静默失效）：**

- **`WrapReqFunc` 会在返回前 `defer handlePushMsgQueue(ctx)` 冲刷推送队列**；`WrapRPCFunc` / 异步类 wrapper **不会**。所以要下发给客户端的 `Push` 只有挂在 `WrapReqFunc` 上才会被自动发出。
- 同步类 wrapper 失败时自动 `ReplyErrorAbort`（回复错误码并 `Abort`）；成功时自动 `Reply`。
- 异步类 wrapper（`WrapAsyncReqFunc` / `WrapAsyncRPCFunc`）**不自动回复**，必须用 `ContextSuspender` 挂起、在回调里手动 `Reply` / `ReplyErrorAbort` / `Resume`（示例见 [player/chat.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/handler/player/chat.go#L17-L29)）。
- `WrapCastFunc` 返回 `false` 时执行 `ctx.Abort()`（[wrapper.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/handler/wrapper.go#L78-L86)），仅用于 `Ntf`。

各场景处理函数示例参考 [player/item.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/handler/player/item.go)（同步）与 [server/server.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/handler/server/server.go#L45-L53)（Cast）。

## 2. 中间件与登录校验

`RegisterC2S` / `RegisterS2S` 接受**变长 `actor.HandlerFunc`**，按顺序排列：前面的作为中间件（可 `ctx.Abort()` 短路），最后一个才是业务 handler。

- 中间件模式：`func mdCheckLogin(ctx *actor.Context)`，未登录直接 `ctx.Abort()`（[player/middleware.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/handler/player/middleware.go)）。
- player 目录提供了便捷封装 `registerC2SFunc(msg, checkLogin bool, f...)`：`checkLogin=true` 时自动前置 `mdCheckLogin`（[player/handler.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/handler/player/handler.go#L34-L40)）。需要登录校验的 c2s 请求统一走它。

```go
// c2s：需要登录校验的同步请求
registerC2SFunc((*pbc2s.UseItemReq)(nil), true, actorhandler.WrapReqFunc(handleUseItem))
// s2s：服务间 RPC
registerS2SFunc((*pbs2s.NotifyJoinChatRoomReq)(nil), actorhandler.WrapRPCNoRespFunc(handleNotifyJoinChatRoom))
```

## 3. 可用的辅助功能

### 3.1 回复（[wrapper.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/handler/wrapper.go#L88-L117)）

- `Reply(ctx, msg)` — 回复任意 pb 消息。
- `ReplySuccess(ctx)` — 回复 `pbcommon.Success`。
- `ReplyError(ctx, err)` — 把 `error` 映射为 `pbcommon.Error`。**只有 `*actor.PbError` 会保留其错误码，其它一律降级为 `ECInternalError`**。
- `ReplyErrorAbort(ctx, err)` — `ReplyError` + `ctx.Abort()`。异步回调里最常用。

### 3.2 错误构造（[error.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/error.go)）

- `actor.WithPbError(code, args...)` → `*actor.PbError`。**协议类错误必须用它构造**，否则 `ReplyError` 只会回 `ECInternalError`。

```go
if req.ItemId == 0 || req.Num <= 0 {
    return nil, actor.WithPbError(pbcommon.ErrCode_ECInvalidPacket)
}
```
（参数校验失败 → `ECInvalidPacket`；领域失败用对应业务错误码，见 `ggs-actor-feature` 的错误码分段。）

### 3.3 参数与推送（[params.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/handler/params.go)）

- `GetArgs[Req](ctx)` — 取出框架已解码的请求消息（`SugarContext(ctx).GetMsg().(Req)`）。wrapper 内部已调用，业务一般不需要直接用。
- `AppendPushMsg(ctx, msg)` — 追加一条 s2c 推送；由 `WrapReqFunc` 的 `defer` 在返回时统一下发。**只在 c2s 同步 `WrapReqFunc` 链路上有效**，其它 wrapper 不会冲刷队列。常配合 systems 使用（如 [systems/item.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/systems/item.go#L43-L45)）。

### 3.4 日志（[log.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/handler/log.go)）

- `handler.Logger()` — 名为 `handler` 的 logger，服务侧常再包一层 `handler.Logger()`。不要在 handler 里临时 `log.New`。

### 3.5 上下文语法糖（[context.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/context.go)）

- `actor.CtxActor[*actors.Player](ctx)` — 取当前 Actor（泛型）。
- `actor.CtxActorModule[*player.Items](ctx, autoCreate)` / `actor.GetActorModule[...]` — 取模块。
- `actor.SugarContext(ctx)` — `.SetMsg` / `.GetMsg` / `.Decode` / `.Reply` / `.RPCWithTimeout` / `.RPC` / `.AsyncRPCWithTimeout` / `.AsyncRPC` / `.Cast`。
- `actor.NewContextSuspender(ctx).Suspend()` → 得到 `*ContextSuspender`，异步回调里 `cs.Resume()`。
- 异步 RPC 结果统一用 `actor.NewRPCResult(reply, err)` 解析：`.Success()` / `.Reply()` / `.Err()` / `.ErrCode()`（[rpc.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/rpc.go)）。
- 服务端主动推送（不排队）：`actor.CActorSugared{CActor: ctx.Actor().(actor.CActor)}.PushRawMessage(msg)`。

## 4. 上下文 key（CtxK）：构造与存取

用于在**一次同步处理链内**给 `*actor.Context` 挂载自定义参数（例如缓存中间结果、标记状态，让后续 handler / 中间件复用）。

### 4.1 构造

```go
var ctxKeyXxx = actor.NewCtxK[V]("xxx") // 返回 actor.CtxK[V]（底层是 kv.K[V]）
```

### 4.2 存取

```go
actor.CtxKSet(ctx, ctxKeyXxx, v)        // 写入
v, ok := actor.CtxKGet(ctx, ctxKeyXxx)  // 读取，返回 (V, bool)
v := actor.CtxKMust(ctx, ctxKeyXxx)     // 必填，缺失 panic
```

### 4.3 硬约束

- **key 必须是包级变量、全局唯一，禁止在使用点临时 `NewCtxK`**。判等基于内部字符串指针的**对象身份**（见 [kv.go](file:///Users/godyy/work/go/godyy/ggs/vendor/github.com/godyy/gutils/container/kv/kv.go#L47-L65)），同名两次 `NewCtxK` 得到的是不同 key，彼此取不到值。
- 泛型 `V` 必须与写入值类型一致；存取不匹配会 **panic**，而不是返回零值 / `false`。
- **作用域仅限当前同步请求链**：`CtxK` **不跨异步边界**。异步回调需要携带参数时用 `CbCtx`（见 `ggs-cbctx`），不要用 `CtxK`。
- key 名称字符串只用于错误信息可读性，不参与判等，仍应语义化命名。

### 4.4 key 落点

| 作用域 | 落点 | 现有示例 |
| --- | --- | --- |
| 通用框架级 | `internal/infra/actor/*.go` 包级 var | `ctxKeyMsg` [context.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/context.go#L43-L43)、`ctxKSeq`（ggskit） |
| handler 包通用 | `internal/infra/actor/handler/params.go` 包级 var | `ctxKeyPushMsgQueue` [params.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/handler/params.go#L8-L10) |
| 仅某用例/文件 | 就近定义在使用文件顶部 | — |
| 服务内跨文件共享 | `app/<service>/internal/...`（不要提到根级） | — |

约定用 `var (...)` 块集中声明，命名以 `ctxKey` 前缀开头，与既有风格一致。

## 5. 检查清单

- 选对了 wrapper：方向（c2s/s2s）+ 同步/异步/单向三重匹配；需要下发 `Push` 时必须是 `WrapReqFunc`。
- 异步 wrapper 已在回调里手动 `Reply` / `ReplyErrorAbort` 并 `cs.Resume()`。
- 业务错误用 `actor.WithPbError(code)` 构造，未用裸 `errors.New` 直接返回（否则降级成 `ECInternalError`）。
- 注册写在 `<actor>/handler.go` 的 `init()`，用消息类型（`(*pb...)(nil)`）而非 PID；c2s 走 `RegisterC2S`，s2s 走 `RegisterS2S`。
- 需要登录校验的 c2s 请求用了 `registerC2SFunc(msg, true, ...)`。
- 新增的 `CtxK` 是包级变量、全局唯一，未在使用点临时 `NewCtxK`；没有把跨异步的参数塞进 `CtxK`。
- handler 内只做协议适配，业务编排落到 `app/<service>/internal/systems`（见 `ggs-actor-feature`）。
- 改了协议先 `make protos`，再 `go build ./... && go test ./...`。
