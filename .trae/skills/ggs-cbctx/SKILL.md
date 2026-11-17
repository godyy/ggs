---
name: ggs-cbctx
description: Build and use actor callback contexts (CbCtx) in ggs — define keys, read params in async callbacks, and release. Use when writing AsyncRPC callbacks or async save/timer handlers.
---

# GGS 回调上下文 (CbCtx) 使用 Playbook

把"异步回调需要携带哪些参数"标准化。CbCtx 是一个键值映射：发起异步调用时携带参数，回调执行时按 key 取回。三条核心原则：**key 必须包级复用**、**回调首行 defer 释放**、**不放宽类型校验**。

## 0. 它是什么

- `actor.CbCtx` 是 [cbctx.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/cbctx.go) 对 `kv.M` 的别名，本质是 `map[any]any` 的封装，带运行期类型校验。
- 生命周期跨异步边界：发起方 `NewCbCtx(...)` 构造 → 作为最后一个 `cbctx any` 参数传给 `AsyncRPC*` / `StartTimer` → 回调里通过 `resp.CbCtx` 取回 → 回调结束前 `CbCtxRelease`。
- 底层由 `sync.Pool` 复用（`gactor.poolOfCbCtx`），所以"必须释放"不是可选优化，而是正确性要求。

不要与请求级上下文的 kv 混淆：`actor.NewCtxK` / `CtxKGet` / `CtxKSet`（见 [context.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/context.go#L13-L29)）只在一个同步请求处理链内有效，不跨异步边界。

## 1. 定义 key

key 类型为 `actor.CbCtxK[V]`，通过 `actor.NewCbCtxK[V]("name")` 创建。

**硬约束**

- 必须是**包级变量**且全局唯一，禁止在使用点临时 `NewCbCtxK`。原因：key 判等基于内部字符串指针的**对象身份**，同名两次 `NewCbCtxK` 得到的是不同的 key，彼此取不到值。
- 泛型 `V` 必须与存入值的类型一致；存取类型不匹配会 **panic**，而不是返回零值或 `false`。
- 名称字符串只用于错误信息可读性，不参与判等，但仍应语义化命名。

**按作用域选择落点**（与仓库分层一致：服务私有 → `app/<service>/internal`，多服务共享 → 根级 `internal/`）

| 作用域 | 落点 | 现有示例 |
| --- | --- | --- |
| 仅某文件/用例一次性使用 | 就近定义在使用的 `.go` 文件顶部 | `cbctxKeySendMsgParams`（[chat.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/systems/chat.go#L242-L242)） |
| 单服务内跨模块共享 | `app/<service>/internal/base/cbctxkey/` | [cbctxkey/cbctx.go](file:///Users/godyy/work/go/godyy/ggs/app/game/internal/base/cbctxkey/cbctx.go) |
| 多服务共享 | 根级 `internal/`（例如 `internal/infra/actor` 下），不要放进某个服务目录 | — |

- 不要把 game 专用 key 提到根级，也不要把多服务通用 key 塞进某个服务目录。
- 通用约定 key：`cbctxkey.Callback = actor.NewCbCtxK[any]("callback")` 承载回调闭包本身；取回后需做函数类型断言。

## 2. 在回调中取参数

**发起方构造**：`actor.NewCbCtx(k1, v1, k2, v2, ...)`，参数按键值交替成对展开，作为 `AsyncRPC*` 的最后一个参数：

```go
if err := p.AsyncRPCWithTimeout(m.MgrUID(p.ServerID()), &pbs2s.SendChatMsgReq{
	RoomId:   roomId,
	SenderId: p.ID(),
	Content:  content,
}, m.onAsyncPlayerSendMsg, actor.NewCbCtx(
	cbctxkey.Callback, cb,
	cbctxkey.Player, p,
	cbctxkey.ChatRoomID, roomId,
), 5*time.Second); err != nil {
	// 同步失败：回调不会被调用，需自行释放 cbctx（见第 3 节）
}
```

**回调取参**：`resp.CbCtx` 是 `any`，先断言为 `*actor.CbCtx`，再按 key 取：

```go
func (m *chatModule) onAsyncPlayerSendMsg(resp actor.ActorRPCResp) {
	cc := resp.CbCtx.(*actor.CbCtx)
	defer actor.CbCtxRelease(cc)

	player := actor.CbCtxMust(cc, cbctxkey.Player)   // 必填：缺失则 panic
	roomId := actor.CbCtxMust(cc, cbctxkey.ChatRoomID)
	cb := actor.CbCtxMust(cc, cbctxkey.Callback).(func(p *actors.Player, err error))
	// ...
}
```

取参选择：

- 必填参数用 `actor.CbCtxMust(cc, key)`，缺失即 panic，尽早暴露接线错误。
- 可选参数用 `actor.CbCtxGet(cc, key)`，返回 `(V, bool)`。
- 取出闭包（`cbctxkey.Callback`）后必须做函数类型断言。

回调中不要保存 `resp.Ctx`（请求 Context）或 `cc` 到回调生命周期之外，回调返回后上下文即失效。

## 3. 释放

- 在回调函数**第一行**写 `defer actor.CbCtxRelease(cc)`。不释放会持续从池中取新对象、并让被引用对象无法及时回收。
- 释放**恰好一次**：不要中途提前手动释放，也不要重复释放。
- **同步失败分支**：`AsyncRPC*` 返回 error 时回调不会执行，框架只释放它自己包装的那层 `kv.M`（见 ggskit `AsyncRPCWithDeadline` 的错误分支），**不会**释放你传入的 `*actor.CbCtx`。此分支必须自行 `actor.CbCtxRelease`。
- 定时器回调：单次（非周期）定时器由框架在触发后自动释放；周期定时器复用同一 cbctx，回调内**不要**释放它。

## 4. 检查清单

- key 是包级变量、全局唯一，没有在使用点临时 `NewCbCtxK`。
- key 落点符合分层：私有就近 / 服务内 `cbctxkey` / 多服务根级 `internal/`。
- `NewCbCtx` 的键值成对，泛型类型与存入值一致。
- 回调首行 `defer actor.CbCtxRelease(cc)`，且只释放一次。
- 必填用 `Must`，可选用 `Get`；闭包做了函数类型断言。
- `AsyncRPC*` 同步失败分支已释放 cbctx。
- 未把请求 Context 或 CbCtx 逃逸到回调之外。
