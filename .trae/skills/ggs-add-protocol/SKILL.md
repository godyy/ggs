---
name: ggs-add-protocol
description: Add or change ggs actor protocols under internal/infra/actor/protocol, covering proto layout (c2s/s2s/common), Req Resp Push Ntf suffixes, per-domain files, and make protos. Use when adding .proto messages, pushes or notifies. Not for handler or system code.
---

# GGS 添加协议 Playbook

把“新增/修改一条 Actor 协议”标准化。协议源文件与生成代码都在 `internal/infra/actor/protocol/`，本技能只负责 proto 结构与代码生成；handler / systems / model 的实现见 `ggs-actor-feature` 技能。

## 0. 目录结构

```text
internal/infra/actor/protocol/
├── protos/
│   ├── c2s/       # 客户端↔服务端（前后端）协议, package c2s
│   ├── s2s/       # 服务↔服务协议, package s2s
│   └── common/    # 通用结构, package common
├── pb/            # 生成代码, 不手工编辑
├── registry/
│   ├── c2s/       # c2s 注册表: c2s.go(手写) + register.go(生成)
│   └── s2s/       # s2s 注册表: s2s.go(手写) + register.go(生成)
├── tools/gen_register/main.go
└── Makefile
```

## 1. 协议结构：后缀决定用途

每条顶层协议消息的名字**必须以固定后缀结尾**，后缀决定它的方向与语义。`tools/gen_register` 只扫描 `Req / Resp / Push / Ntf` 四种后缀（见 [main.go](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/protocol/tools/gen_register/main.go#L24-L24)），**其它名字不会被注册，也就无法编解码收发**。

| 方向 | 语义 | 后缀 | 示例 | 对应 handler 包装 |
| --- | --- | --- | --- | --- |
| c2s | 客户端请求（同步） | `Req` | `UseItemReq` | `WrapReqFunc` |
| c2s | 客户端请求（异步） | `Req` | `SendChatMsgReq` | `WrapAsyncReqFunc` |
| c2s | 服务端 → 客户端 **推送**（s2c） | `Push` | `ItemPush` `ChatMsgPush` | 服务端主动 send/broadcast，无 handler |
| c2s | 请求响应 | `Resp` | `UseItemResp` | — |
| s2s | 服务间 RPC 请求 | `Req` | `NotifyJoinChatRoomReq` | `WrapRPCFunc` / `WrapRPCNoRespFunc` |
| s2s | 服务间 **通知/Cast**（单向、无响应） | `Ntf` | `PlayerOnlineNtf` `ActorSaveResultNtf` | `WrapCastFunc` |
| s2s | RPC 响应 | `Resp` | `ChatHistoryResp` | — |

关键点：**s2c 推送用 `Push`，服务间推送用 `Ntf`，两者后缀不同，不要混用。**

反例（不会被注册，禁止）：`XxxNotify`、`XxxNotify2Client`、`XxxResponse`、`XxxMsg`。

结构与示例：

```protobuf
// 请求: 玩家使用道具
message UseItemReq {
  int32 itemId = 1; // 道具ID
  int64 num = 2;    // 使用数量
}

// 响应: 与请求配对
message UseItemResp {
  int32 itemId = 1;  // 道具ID
  int64 leftNum = 2; // 剩余数量
}

// s2c 推送: 服务端主动下发
message ItemPush {
  repeated common.Item items = 1; // 道具变更
}

// s2s 通知: 服务间单向 cast
message PlayerOnlineNtf {}
```

## 2. 三层划分：前后端 / 服务间 / 通用

- **c2s**（`protos/c2s/`）：面向前端的请求/响应/推送。`package c2s`。
- **s2s**（`protos/s2s/`）：服务内部或服务间的通信（Actor 间 RPC/Cast）。`package s2s`。
- **common**（`protos/common/`）：被 c2s 与 s2s **共同引用的嵌套结构**，如 `ChatMsg`、`Item`、`PlayerBase`；以及全局 `ErrCode` 枚举、`Error`、`Success`。`package common`。

判断规则：

- 能被两个方向复用的**嵌套结构** → `common`。
- 带方向语义的**顶层包**（`Req`/`Resp`/`Push`/`Ntf`）→ `c2s` 或 `s2s`。
- 只有 `common.Error` 与 `common.Success` 会被注册进 c2s、s2s 两个 registry（见 `commonMessageRegistrations`）；其它 common 结构只作为字段嵌套，不单独注册。

每个 proto 文件必须使用与目录一致的头部：

```protobuf
syntax = "proto3";

package c2s; // 或 s2s / common

import "common/xxx.proto"; // 需要引用通用结构时

option go_package = "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s";
```

`go_package` 必须与所在目录对应（c2s/s2s/common 三选一），import 路径以 `common/` 等相对 protos 根，不能写全路径。

## 3. 按功能系统划分 proto 文件

- **一个功能域 = 一个 proto 文件，文件名即域名**：`chat.proto`、`item.proto`、`player.proto`、`login.proto`。
- 同一功能域若同时有前后端与服务间协议，在 `c2s/` 与 `s2s/` 各放一份**同名文件**（如 `c2s/chat.proto`、`s2s/chat.proto`）。
- 复用的嵌套结构放 `common/` 下的同名文件（如 `common/chat.proto`）。
- 错误码统一维护在 `common/error.proto` 的 `ErrCode` 枚举里，**按功能域分段**并写注释：

```protobuf
enum ErrCode {
  ECSuccess = 0; // 成功.

  // 道具系统 [1000, 1999]
  ECItemNotEnough = 1000; // 道具数量不足.

  // 聊天系统 [2000, 2999]
  ECChatRoomNotExists = 2000; // 聊天室不存在.
}
```

新增一个功能域（以 `mail` 为例）时的完整改动面：

1. `protos/common/mail.proto` — 复用嵌套结构（如有）。
2. `protos/c2s/mail.proto` — 前后端 `Req`/`Resp`/`Push`。
3. `protos/s2s/mail.proto` — 服务间 `Req`/`Resp`/`Ntf`（如有）。
4. `protos/common/error.proto` — 新增该域的错误码段。
5. 执行代码生成（见第 4 节）。

避免把不同功能域塞进同一个 proto 文件，避免单文件膨胀。

## 4. 生成协议代码

在仓库根目录执行：

```bash
make protos
```

它会 `cd internal/infra/actor/protocol` 并执行该目录的 [Makefile](file:///Users/godyy/work/go/godyy/ggs/internal/infra/actor/protocol/Makefile)，依次完成：

1. `clang-format` 格式化 `protos/**/*.proto`。
2. 清空重建 `pb/`，用 `protoc` 从 `protos/` 生成 `pb/**`。
3. 运行 `tools/gen_register/main.go` 扫描 c2s/s2s 的 proto，重新生成 `registry/c2s/register.go`、`registry/s2s/register.go`。
4. `goimports` 格式化 registry。

验证：

```bash
go build ./...
go test ./...
```

## 5. 硬约束与检查清单

- 不手工编辑 `pb/**` 与 `registry/*/register.go`（生成物，含 `DO NOT EDIT`）。只改 `.proto`，再 `make protos`。
- 消息名必须以 `Req` / `Resp` / `Push` / `Ntf` 结尾，否则不会被注册。
- s2c 推送用 `Push`，服务间推送用 `Ntf`，不要混用。
- `package` 与 `go_package` 必须和 proto 所在目录（c2s/s2s/common）一致。
- 通用可复用的嵌套结构放 `common`；顶层 Req/Resp/Push/Ntf 放对应方向目录。
- 新功能域独立成文件，错误码在 `common/error.proto` 按段追加。
- 协议改动完成后，记得同步在 handler 层注册（`registerC2SFunc` / `registerS2SFunc`），详见 `ggs-actor-feature` 技能。
