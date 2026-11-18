# ggs

多服务 Go 游戏后端，模块名 `github.com/godyy/ggs`，Go 版本 `1.25.0`。

本仓库基于自研 Actor 运行时 `gactor` 与通用工具包 `ggskit` 构建，目标形态是「单玩家强状态、弱锁、可横向扩缩容」的游戏服务器。

> 详细的架构分析见 [服务器架构分析报告.md](docs/服务器架构分析报告.md)；面向 AI/协作的开发约束见 [AGENTS.md](AGENTS.md)。

---

## 服务

| 服务 | 类型 | 职责 |
| --- | --- | --- |
| `app/agent` | 网关 | 校验登录 token，选择 game 节点，将客户端请求转发到玩家 Actor；通过 `ggskit/infra/cluster` 参与集群 |
| `app/game` | 运行时 | 承载游戏核心业务 Actor（`Server` / `Player`），将项目 Actor 定义装配到 `ggskit/infra/actor.Service` |
| `app/login` | HTTP | 角色列表 / 创建 / 登录与登录 token 签发；使用 Redis / Mongo / actor registry / `ServerStore`，但不加入集群 |
| `app/platform` | HTTP | 平台业务，目前实现创建服务元数据并预注册固定 `Server` Actor |
| `app/client` | CLI | 本地客户端 / 机器人工具，支持多级自动补全与 protobuf JSON 输入 |

---

## 架构分层

```
底层框架（vendor，通用）     gactor / gnet / gcluster / gmpsc / gtimewheel / gevent / grendezvous / glog
        │
应用工具包（可复用）         ggskit  base(基础能力) / infra(基础设施) / utils(纯函数工具)
        │
项目共享层（本项目共用）      internal/  base / infra/actor / infra/mongo / gdconf / protocol ...
        │
服务层（各自接线）            app/<service>/internal
```

判断规则：服务专属策略放 `app/<service>/internal`；多服务共享但只属于本项目放根级 `internal/`；可复用于多个游戏项目的通用能力放 `ggskit`。

---

## 目录结构

```text
app/
├── agent/           # 客户端网关
├── game/            # 游戏运行时
├── login/           # HTTP 登录 / 角色服务
├── platform/        # HTTP 平台服务
└── client/          # 本地客户端 / 机器人
internal/
├── base/            # 项目常量、生命周期、logger、packet、nodeutil
├── infra/
│   ├── actor/       # 项目 Actor 基座：泛型基类、语法糖、CtxK/CbCtx、持久化、migration
│   ├── actors/      # 项目 Actor 类别、Define、模型
│   ├── mongo/       # 共享 Mongo 模型与 repo
│   ├── mongobd/     # 后台批量写盘封装
│   └── monitor/     # 监控 / 探针适配
├── models/          # 共享业务模型（如登录 token 载荷）
├── gdconf/          # 导表生成配置（MongoDB 数据源）
├── protocol/        # .proto 源、生成代码、注册表、工具
└── utils/           # 共享工具（Gin / HTTP helper）
```

---

## 核心机制

- **Actor 运行时**：每个 Actor 一个独立 goroutine，内部状态天然串行化。邮箱按**优先级分层**，消息语义分 **RPC**（请求/响应）、**Cast**（单向）、**Forward**（转发）。
- **持久化**：内存态修改打**脏标记** → `DelaySave` 延迟保存 → 落盘走 `mongobd` 后台 worker 池批量写，避免阻塞 Actor 主循环。
- **Actor 路由**：固定节点策略 `ActorFixedNode` **优先于**分组路由 `ActorNodeGroup`；未注册 Actor 由 `grendezvous` 一致性哈希选择节点。
- **集群**：`gcluster` 提供节点间长连接通信，服务发现由 etcd（`ggskit/infra/cluster`）负责。
- **Actor 迁移**：三阶段（接管 → 转发 → 清理），规划函数基于**冻结计划**、租约保证幂等，支持动态扩容/缩容。
- **泛型**：仅作为边界处的**类型安全外壳**（handler 包装、模型基类、`CtxK`/`CbCtx`），运行时核心保持接口化，对性能无实质影响。

---

## 依赖

主要 Go 依赖（均以 vendor 方式引入）：

- [gactor](https://github.com/godyy/gactor) `v0.4.17` — Actor 运行时
- [ggskit](https://github.com/godyy/ggskit) `v0.2.16` — 通用工具包
- `gcluster` / `gnet` / `gmpsc` / `gtimewheel` / `gevent` / `grendezvous` / `glog`

本地中间件依赖：

- Redis
- MongoDB
- etcd

---

## 快速开始

启动本地依赖：

```bash
docker compose -f docker/dev/docker-compose.yaml up -d
```

启动服务：

```bash
go run ./app/platform
go run ./app/login
go run ./app/game -env-server-id=1
go run ./app/agent -env-server-id=1
make run_client
```

配置：

- 通过 `-config-path` 指定 TOML 配置。
- `agent` 与 `game` 要求 `-env-server-id`。
- `game` 派生 DB 名：`game_<serverId>`。

---

## 测试

```bash
go test ./...
go test ./app/agent/internal/infra/router ./internal/infra/mongobd
```

修改协议后：

```bash
make protos
go test ./...
```

修改导表配置后：

```bash
make gen_gdconf
```

---

## 文档

- [AGENTS.md](AGENTS.md) — 目录边界、硬约束、常见改动落点、风格约定
- [服务器架构分析报告.md](docs/服务器架构分析报告.md) — 底层框架、应用层组件、泛型与性能分析
- [服务器架构发展建议.md](docs/服务器架构发展建议.md) — 后续迭代路线图与实施清单
