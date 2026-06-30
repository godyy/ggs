# AGENTS.md

## 1. 核心定位

本仓库是具体的多服务 Go 游戏后端，模块名：`github.com/godyy/ggs`，Go 版本：`1.25.0`。

通用游戏服工具已经抽离到 `github.com/godyy/ggskit`。当前仓库只应承载：

- 各服务启动与接线
- 项目共享 Actor 定义、模型、协议
- 具体业务 handler / system / repo
- 针对本项目的轻量封装

不要在本仓库重新实现 `ggskit` 已拥有的通用能力，例如 actor registry、actor router、cluster service、DB client、protocol registry 基础能力等。

目前项目中，`server`和`game`服务之间是简单的一一对应关系，即每个`server`有且仅有一个`game`节点与之对应。后期会根据项目实际情况和需求，做出调整。

---

## 2. 目录边界

### 2.1 服务层

- `app/agent`
  - 客户端网关。
  - 校验登录 token，选择 game 节点，把客户端请求转发到玩家 Actor。
  - 通过 `ggskit/infra/cluster` 参与集群。

- `app/game`
  - 游戏运行时。
  - 承载游戏核心业务逻辑相关 Actor, 目前暂时简单定义了`Server`和`Player`。
  - 将项目 Actor 定义装配到 `ggskit/infra/actor.Service`。

- `app/login`
  - HTTP 登录 / 角色服务。
  - 负责角色列表、创建、登录与登录 token 签发。
  - 使用 Redis / Mongo / actor registry / `ServerStore`，但不加入 cluster, 可以看作是支持负载均衡的微服务。

- `app/platform`
  - HTTP 平台服务, 对接平台相关业务。
  - 目前只简单实现了创建服务元数据的业务，预注册固定节点 `Server` Actor，后续会根据需求扩展其他业务。

- `app/client`
  - 本地客户端 / 机器人工具。
  - 交互式 CLI 基于 `github.com/chzyer/readline`，支持多级自动补全（命令/参数）。
  - protobuf JSON 输入/模板使用 `google.golang.org/protobuf/encoding/protojson`（例如 `sendreq`），可通过 `EmitUnpopulated` 输出 0 值字段。

各服务自己的接线、handler、配置封装、服务私有 infra 放在：

```text
app/<service>/internal
```

### 2.2 根级共享层

旧的共享 `app/internal` 已迁移到根级 `internal/`。根级 `internal/` 是项目共享层，不是通用基础设施层。

- `internal/base`
  - 项目常量、生命周期 hook、logger 初始化、packet helper、nodeutil。

- `internal/infra/actors`
  - 项目 Actor 类别、Define、模型、持久化接入。

- `internal/infra/mongo`
  - 项目共享 Mongo 模型与 repo。

- `internal/infra/mongobd`
  - 对 `ggskit/infra/mongobd` 的项目封装，增加后台消费者处理。

- `internal/infra/monitor`
  - 监控 / 探针的轻量路由适配。

- `internal/models`
  - 项目共享业务模型，例如登录 token 载荷。

- `internal/gdconf`
  - 由导表工具（`gexcels`）生成的游戏配置代码与加载逻辑（以 MongoDB 为数据源）。
  - 约定：除 `*_ext.go` 外，其余文件均为生成物，不要手工编辑；需要变更时走重新生成流程。

- `internal/protocol`
  - `.proto` 源文件、生成代码、协议注册表、生成工具。

- `internal/utils`
  - 项目共享工具，例如 Gin / HTTP helper。

### 2.3 `ggskit` 边界

以下能力属于 `ggskit`：

- `actor.Registry` / `actor.ServerStore` / `actor.Router`
- `actor.Service` / `actor.Client` / `actor.Codec`
- `cluster.Service`
- Redis / Mongo client
- env / flags / logger / auth / codec / protocol 基础能力
- 通用 monitor / probe 基础能力

判断规则：

- 服务专属策略：放在 `app/<service>/internal`。
- 多服务共享但只属于本项目：放在根级 `internal/`。
- 可复用于多个游戏项目的通用能力：放在 `ggskit`。

---

## 3. 硬约束

### 3.1 节点命名

必须保持：

```text
NodeName = strconv.FormatInt(ServerId, 10)
NodeID   = category/nodeName
示例      = game/101
```

规则：

- `NodeName` 必须且只能由 `ServerId` 推导。
- 不要重新引入手工配置 `NodeName`。
- 服务节点构造优先使用：
  - `internal/base/nodeutil.MakeServerNodeName(serverId)`
  - `internal/base/nodeutil.NewServerNode(category, serverId, addr)`

### 3.2 Actor 位置与归属

必须区分：

- `Registry`：回答 Actor 当前在哪里。
- `ServerStore`：回答 Actor 归属于哪个服务器。

规则：

- `Registry` 只负责运行时位置。
- `ServerStore` 只负责稳定归属关系。
- 不要给 `ServerStore` 增加 TTL / lease 语义。
- 不要把 `ServerStore` 当作节点可用性来源。
- Redis client 缺失属于接线错误，应在构造期 fail-fast。

当前接入点：

- `app/login/internal/app/actor.go`：构造 registry 与 server store。
- `app/login/internal/handlers/character.go`：角色创建后写入玩家归属。
- `app/game/internal/app/actor.go`：fallback 路由时从 `ServerStore` 读取玩家归属。
- `app/platform/internal/app/actor.go`：构造 registry，用于预注册固定 `Server` Actor。

### 3.3 Actor 路由优先级

路由优先级固定为：

1. `ActorFixedNode`
2. `ActorNodeGroup`

不要改变这个顺序。

本仓库负责的路由策略主要在：

- `app/game/internal/app/actor.go`
  - `getNodeGroup`
  - `getActorFixedNode`
  - `getActorNodeGroup`
  - `getActorServerID`

- `app/agent/internal/infra/router/nodeselector.go`
  - 网关侧 game 节点选择逻辑。

规则：

- 固定节点策略必须优先于分组路由。
- 服务特有路由策略放在服务接线或回调中。
- 不要在本仓库伪造通用 router / selector 抽象。
- 不要重建与底层 selector / router 重复的缓存，除非有明确必要。

### 3.4 构造期校验

项目约定：

- 必填依赖优先在构造阶段校验。
- 缺少必填依赖时返回 error，尽早失败。
- 不要用运行期 nil guard 掩盖错误接线。
- 构造期非法参数错误倾向于直接 `errors.New`。
- 需要嵌套上下文 error 时，优先使用 `pkgerrors.WithMessage` / `WithMessagef`，但不要过度包装简单透传错误。

---

## 4. 关键流程

### 4.1 角色登录链路

1. 客户端调用 `login` HTTP API 获取登录 token。
2. 客户端连接 `agent`，发送 `LoginReq`。
3. `agent` 校验 token，并获取分布式登录锁。
4. `agent` 查询 actor registry，获取玩家 Actor 位置。
5. 若位置缺失或失效，`agent` 基于 `serverId + playerId` 选择 game 节点。
6. `agent` 更新玩家 Actor 位置到 registry。
7. `agent` 连接玩家 Actor，转发登录 / 玩法流量。
8. `game` 初始化玩家状态并返回响应。

### 4.2 平台创建服务器链路

1. `platform` 在 MongoDB 创建服务器记录。
2. 从 `game/<serverId>` 推导 `NodeId`。
3. 在 registry 中预注册对应的 `Server` Actor。

`Server` Actor 是固定节点 Actor，其归属由 server ID 直接推导，不依赖 `ServerStore`。

---

## 5. 常见改动落点

### 5.1 新增玩家玩法消息

- Proto：`internal/protocol/protos/c2s` 或 `internal/protocol/protos/s2s`
- Game handler：`app/game/internal/handlers/player`
- 领域编排：`app/game/internal/systems`
- 共享 Actor / 类别 / 模型：`internal/infra/actors`

修改 proto 后执行：

```bash
make protos
```

### 5.2 新增 login / platform HTTP API

- Handler：`app/<service>/internal/handlers`
- 路由注册：沿用该服务现有 handler init/setup 模式
- 服务私有 repo：`app/<service>/internal/infra/repo`
- 多服务共享模型 / repo：`internal/infra/mongo`

### 5.3 修改 Actor 路由行为

优先级：

1. 先改 `app/game/internal/app/actor.go` 的服务级回调。
2. 再看 `app/agent/internal/infra/router/nodeselector.go` 的网关选择逻辑。
3. 只有确实属于通用能力时，才改 `ggskit`。

### 5.4 修改 cluster 行为

- 服务接线：
  - `app/game/internal/app/cluster.go`
  - `app/agent/internal/app/cluster.go`

- 通用 cluster 内核：改 `ggskit`，不要在本仓库本地化实现。

### 5.5 修改节点构造

优先只改：

```text
internal/base/nodeutil
```

这是高风险区域，必须保持 `ServerId -> NodeName -> NodeID` 约定。

### 5.6 修改 Actor 归属映射

- login 写入路径：`app/login/internal/handlers/character.go`
- game 读取与 fallback 路由：`app/game/internal/app/actor.go`
- 服务接线：`app/<service>/internal/app/actor.go`
- `ServerStore` 通用语义：属于 `ggskit`

如果归属来源变化，写入路径和 fallback 读取路径必须一起更新。

### 5.7 修改登录 token 共享载荷

- 共享载荷定义：`internal/models`
- token 生成 / 校验使用：各服务 handler 或 hook
- 修改时保持 `agent` 预期的 token 结构兼容。

### 5.8 修改导表配置（gdconf）

- 代码位置：`internal/gdconf`
- 数据源：MongoDB（game 启动时会加载）
- 修改流程：优先修改导表源（默认 `../ggs_excels`），然后执行：

```bash
make gen_gdconf
```

---

## 6. 协议与生成代码

协议目录：

```text
internal/protocol
├── protos      # 可编辑 .proto 源文件
├── pb          # 生成代码，不手工编辑
├── registry    # 协议注册表，部分文件为生成物
└── tools       # 生成工具
```

Actor 协议（C2S/S2S）目录：

```text
internal/infra/actor/protocol
├── protos      # 可编辑 .proto 源文件
├── pb          # 生成代码，不手工编辑
├── registry
│   ├── c2s     # C2S 注册表（包含生成的 register.go）
│   └── s2s     # S2S 注册表（包含生成的 register.go）
└── tools       # 生成工具（gen_register）
```

规则：

- 修改 `.proto` 源文件，不直接改 `.pb.go`。
- 不直接改带有 `Code generated ... DO NOT EDIT.` 的文件。
- 修改协议后执行 `make protos`。
- 保持 C2S / S2S 注册与 proto 定义一致。

### gdconf（导表生成代码）

目录：`internal/gdconf`

规则：

- 文件头包含 `Code generated by gexcels; DO NOT EDIT.` 的文件不要手工编辑。
- 允许手工扩展的代码只放在 `*_ext.go`。
- 需要更新时，执行 `make gen_gdconf` 重新生成并格式化代码。

### gdconf 加载后处理函数（AfterLoad）

用途：在配置表加载完成后，构建二级索引 / 派生数据（例如按等级分组、按 id 建反查表等），避免运行期反复扫描。

约定：

- 只能写在 `internal/gdconf/*_ext.go` 中。
- 通过 `RegisterAfterLoadFunc(tableName, func() error { ... }, priority)` 注册。
- 会在 `gdconf.Load(...)` 完成所有表加载后执行；如果走 `gdconf.LoadTable(...)`，只会执行被加载表的 after-load（按 `priority` 从小到大执行）。

推荐写法：

```go
package gdconf

import "sync/atomic"

var levelOfItems atomic.Value

func init() {
	RegisterAfterLoadFunc(TblNameItem, func() error {
		v := make(map[int32][]*Item)
		for _, item := range TblItem().All() {
			v[item.Level] = append(v[item.Level], item)
		}
		levelOfItems.Store(v)
		return nil
	}, 0)
}

func GetLevelOfItems(level int32) []*Item {
	return levelOfItems.Load().(map[int32][]*Item)[level]
}
```

注意点：

- after-load 里只做“纯计算 + 存结果”，不要做网络 IO 或依赖不稳定外部资源。
- 数据结构需并发安全：推荐 `atomic.Value` / `atomic.Pointer` 持有只读 map/slice，写一次后只读访问。
- `priority` 越小越先执行；当多个表之间存在派生依赖时，用 `priority` 明确顺序。

---

## 7. 配置与启动

配置模式：

- 各服务配置结构：`app/<service>/internal/base/config`
- 服务 env 封装：`app/<service>/internal/base/env`
- 通用 flags / env 基础能力：来自 `ggskit`

约定：

- `-config-path` 指定 TOML 配置。
- `agent` 和 `game` 要求 `env-server-id`。
- `game` 派生 DB 名：`game_<serverId>`。
- registry / `ServerStore` 复用服务现有 Redis client，不新增重复客户端或配置源。

---

## 8. 验证建议

优先做贴近改动区域的验证。

当前本仓库较有价值的测试区域：

- `app/agent/internal/infra/router/*_test.go`
- `internal/infra/mongobd/*_test.go`
- `app/game/internal/base/errs/*_test.go`

常用命令：

```bash
go test ./...
go test ./app/agent/internal/infra/router ./internal/infra/mongobd
```

如果改了协议：

```bash
make protos
go test ./...
```

如果改了接线或启动逻辑，应验证对应应用能用 dev 配置和预期 flag 启动。

本地依赖：

- Redis
- MongoDB
- etcd

本地环境：

```bash
docker compose -f docker/dev/docker-compose.yaml up -d
```

常用启动命令：

```bash
go run ./app/platform
go run ./app/login
go run ./app/game -env-server-id=1
go run ./app/agent -env-server-id=1
make run_client
```

---

## 9. 高风险区域

以下区域不要轻易修改：

- `internal/base/nodeutil`
- `internal/infra/actors`
- `app/game/internal/app/actor.go`
- `app/game/internal/app/cluster.go`
- `app/agent/internal/infra/router/nodeselector.go`
- `internal/protocol/pb`
- `internal/protocol/registry/*_register.go`

风险原因：

- 影响多个服务共享行为。
- 承载节点身份、路由、Actor 定义或协议兼容性。
- 小改动可能破坏启动、路由、注册或协议解码。

如果必须修改：

- 控制改动范围。
- 保持既有不变量。
- 增加贴近改动点的验证。

---

## 10. 风格约定

- 保持共享层薄，不要过度设计。
- 命名追求精确且精简。
- 优先直接数据流，避免不必要抽象。
- 通用策略放 `ggskit`，项目策略放本仓库。
- 服务业务优先放服务内，不要过早上提到根级 `internal/`。
- 不要把服务特有策略塞进通用基础设施。
- 不要用宽泛运行期 fallback 掩盖接线错误。
- 如果下层已经持有事实来源，不要重复缓存状态。

---

## 11. 完成前检查

提交或结束任务前确认：

- 改动落点是否正确：服务层、根级 `internal/`，还是 `ggskit`？
- 是否保持 `NodeName = ServerId` 推导规则？
- 是否保持固定节点优先于分组路由？
- 是否保持 `Registry` 与 `ServerStore` 职责分离？
- 是否避免直接修改生成协议文件？
- 如果改了接线，是否验证启动或相关集成行为？
- 如果改动看起来通用，是否确认它是否应该放到 `ggskit`？
- 是否避免新增宽泛运行期兜底逻辑？
