---
name: ggs-actor-feature
description: Playbook for adding an Actor-based feature, covering model modules, systems, c2s/s2s protos, handler wiring, regeneration and tests. Invoke when adding or modifying actor features or protocols.
---

# GGS Actor Feature Playbook

将“新增一个 Actor 能力”标准化为可复用流程，适用于本仓库内所有基于 Actor 模型的服务与 Actor（目前主要是 `game`，后续可扩展更多服务与更多 Actor 类型）。核心原则：数据落在 Model/Module，业务编排落在 systems，协议落在 proto（c2s/s2s），handler 仅做协议适配与错误码映射。

## 0. 目标与边界

- **Module（Model 模块）**：只做状态与状态操作（增删改查 + SetDirty），不做协议、不做跨模块编排、不做外部 IO。
- **systems（系统/用例层）**：聚合多个模块、实现领域动作与规则；输入是 Actor + 领域参数；输出是领域结果（不返回 pb）。
- **handlers（协议层）**：只做“pb → 领域参数”的转换、调用 systems、再把结果/错误映射回“领域结果 → pb”。
- **protocol（协议层）**：proto 按功能拆文件；消息后缀决定方向与语义（`Req`/`Resp`/`Push`/`Ntf`）；错误码按“功能分段”预留区间；变更后必须重新生成。协议细节见 `ggs-add-protocol` 技能。

## 0.1 协议类型选择（c2s vs s2s）

- **c2s**：面向客户端的请求/响应（例如 Player 对接前端的玩法协议）。
- **s2s**：服务内/服务间通信（例如 Server Actor 或跨服务协作协议）。
- **约定（当前项目语义）**
  - Player：主要处理 c2s；当需要服务内/服务间协作时，也可能处理 s2s。
  - Server：只处理 s2s（不直接对接前端 c2s）。
  - 其他 Actor：根据职责决定是否暴露 c2s，是否需要 s2s；优先最小暴露面。
- **后缀约定**：`Req` 请求、`Resp` 响应、`Push` s2c 推送、`Ntf` 服务间通知/Cast；名字必须以四者之一结尾，否则不会被注册。

## 1. 开发流程（推荐顺序）

### 1.1 新增/修改 Model 模块（数据）

1) 在 `internal/infra/actors/models/<actor>/` 下新增模块文件（例如 `items.go`）。
2) 模块结构体嵌入 `moduleBase[*T]`，实现 `ModuleKey() string`。
3) 所有会改变模块数据的方法必须调用 `SetDirty()`。
4) 在 `internal/infra/actors/models/<actor>/modules.go` 注册模块：
   - `actor.RegisterModule[*YourModule](moduleRegistry)`

**检查项**
- 模块字段可被 bson 持久化（必要时自定义 BSON 编解码）。
- `autoCreate=true` 获取模块时不会 panic（ModuleKey 与注册一致）。

### 1.2 新增/修改 systems（业务动作）

1) 在 `app/<service>/internal/systems/` 下按域新增文件（例如 `item.go`、`mail.go`）。
2) 导出一个全局入口变量（保持风格一致）：
   - `var Items = &itemsModule{}`
3) systems 方法签名建议：
   - 输入：`a *actors.<Actor>`（或对应 Actor） + 强类型参数
   - 输出：领域结构体/基础类型 + `ok bool` 或 `error`

**建议**
- 协议参数合法性（如 0/负数）可以在 handler 拦截；领域规则（如数量不足/状态不允许）放 systems。
- systems 内部通过 `actors.GetModule[*xxx](p, true)` 访问模块。

### 1.3 新增/修改协议（proto + error code）

1) 选择协议类型：
   - 面向前端：`internal/infra/actor/protocol/protos/c2s/`
   - 服务内/服务间：`internal/infra/actor/protocol/protos/s2s/`
   - 通用嵌套结构：`internal/infra/actor/protocol/protos/common/`
2) 按域新增 proto 文件（例如 `item.proto`），消息名必须以 `Req`/`Resp`/`Push`/`Ntf` 结尾（s2c 推送用 `Push`，服务间推送用 `Ntf`）。
3) 在 `common/error.proto` 的 `ErrCode` 中为该域分配专属错误码段并写注释：
   - 示例：道具错误码 `[1000,1999]`，从 `1000` 开始递增
4) 生成代码：
   - `make protos`

**硬约束**
- 不直接修改 `internal/infra/actor/protocol/pb/**` 与 `internal/infra/actor/protocol/registry/{c2s,s2s}/register.go`（生成物）。
- 协议结构、分层、生成流程的完整说明见 `ggs-add-protocol` 技能。

### 1.4 新增/修改 handler（协议适配）

1) 在 `app/<service>/internal/handler/<actor>/` 下按域新增 handler 文件（例如 `item.go`），注册入口在同目录 `handler.go` 的 `init`。
2) handler 只做：
   - pb 参数校验（非法 → `ECInvalidPacket`）
   - 调用 `systems.<Domain>.<Func>(...)`
   - 将领域失败映射为 pb 错误码（例如数量不足 → `ECItemNotEnough`）
   - 组装并返回 pb resp
3) 在 handler 的注册文件（`init`）中按消息类型登记（不是按 PID）：
   - c2s：`registerC2SFunc((*pbc2s.YourReq)(nil), checkLogin, actorhandler.WrapReqFunc(handleYour))`；异步请求用 `WrapAsyncReqFunc`。
   - s2s：`registerS2SFunc((*pbs2s.YourReq)(nil), actorhandler.WrapRPCFunc(handleYour))`；无响应用 `WrapRPCNoRespFunc`，单向通知 `Ntf` 用 `WrapCastFunc`。

## 2. 消息后缀与错误码分段

### 2.1 消息后缀（决定方向与语义）

- `Req`：请求（c2s 客户端请求 / s2s RPC 请求）。
- `Resp`：响应，与 `Req` 配对。
- `Push`：s2c 推送，服务端主动下发给客户端（仅 c2s）。
- `Ntf`：服务间通知/Cast，单向无响应（仅 s2s）。
- 只有这四种后缀会被注册；不存在 `pid.proto`，PID 由消息全名哈希生成，无需手工分配（见 `ggs-add-protocol`）。

### 2.2 ErrCode（common/error.proto）

- 在 `common/error.proto` 的 `ErrCode` 枚举中按功能域分段并写注释（示例）：
  - `[0, 999]` 基础/通用
  - `[1000, 1999]` 道具
  - `[2000, 2999]` 聊天

## 3. 最小验证闭环

1) `make protos`
2) `go test ./...`
3) （可选）用 `make run_client` 走一遍相关请求，核对返回码与字段。

## 4. 输出要求（给实现者/AI）

- 变更应按职责落点拆文件：proto/handler/systems/module 不互相污染。
- handler 文件与 proto 文件都按“功能域”拆分，避免出现单文件膨胀。
- 说明变更时给出涉及的文件路径与关键逻辑变化点。
