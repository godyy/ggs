# 动态迁移端到端测试用例

## 用例信息

- 用例名称：Player Actor 从 `game-worker/w1` 动态迁移至 `game-worker/w2`
- 目标：验证迁移期间的 worker 接入隔离、Player 状态迁移、已有连接转发和迁移收尾。
- 覆盖范围：`standby -> receiving -> serving`、源 `serving -> draining`、agent 选点、gactor 迁移收敛、`finish-forward` 和 `finalize`。

## 前置条件

1. 构建 game、agent、login 镜像并启动 Docker 开发环境：

   ```bash
   make docker-game docker-agent docker-login
   docker compose -f docker/dev/docker-compose.yaml up -d
   ```

2. 固定 game、`w1`、`w2`、`w3`、`w4` 均运行于 `ggs_network`，在容器内统一使用 `20002` 集群端口；worker 宿主机迁移管理端口分别为 `22089`、`22090`、`22091`、`22092`。`w4` 复用 `w2` 的配置，初始状态为 `standby`。

3. 等待三个 worker 注册并检查 Redis 状态：

   ```bash
   redis-cli HGET actor:migration:worker:game-worker/w1 state
   redis-cli HGET actor:migration:worker:game-worker/w2 state
   redis-cli HGET actor:migration:worker:game-worker/w3 state
   redis-cli HGET actor:migration:worker:game-worker/w4 state
   ```

   预期依次为 `serving`、`standby`、`serving`、`standby`，且四个 Hash 均包含非空的 `instance_id`、`build_version`、`protocol_version`。

4. 使用两个独立账号创建并登录角色。玩家 A 保持客户端连接和周期性请求，用作迁移对象；玩家 B 在迁移启动后登录，用作 agent 新玩家选点验证。

当前多目标接入依赖 `vendor/github.com/godyy/ggskit/infra/actor/service.go` 中新增的 `StartMigrationWithPlanner` 透传。Docker 使用 vendor 构建；重新生成 vendor 前须先将该接口合入并发布到 ggskit，否则补丁会丢失。

## 测试步骤

| 步骤 | 操作 | 预期结果 |
| --- | --- | --- |
| 1 | 登录玩家 A，记录 agent 日志中的 `pick player worker` 和玩家 A 的 Actor Registry 位置。 | 玩家 A 被分配到 `game-worker/w1` 或 `game-worker/w3`。若不是 `w1`，用新的测试账号重新登录，直到玩家 A 位于 `w1`；测试只迁移 `w1` 上已有的玩家。 |
| 2 | 玩家 A 保持连接，持续发送可观测的 C2S 请求，例如每秒发送一次心跳或已有的无副作用请求。记录迁移前最后一条成功响应。 | Player 已登录且迁移期间具备持续流量。 |
| 3 | 从 `w1` 发起迁移： | HTTP 返回 `202 Accepted`，响应中包含指向 `game-worker/w2` 的迁移快照。 |

```bash
curl -i -X POST http://127.0.0.1:22089/admin/migration/start \
  -H 'Authorization: Bearer dev-migration-token' \
  -H 'Content-Type: application/json' \
  -d '{"targetNodeId":"game-worker/w2"}'
```

| 步骤 | 操作 | 预期结果 |
| --- | --- | --- |
| 4 | 在收到步骤 3 响应后立即读取三个 Redis 状态。 | `w1=draining`、`w2=receiving`、`w3=serving`。`w2` 的 `migration_id` 非空；`w1` 不再可被 agent 用于新玩家。 |
| 5 | 在 `w2` 仍为 `receiving` 时登录玩家 B，查看 agent 的 `pick player worker` 日志及 B 的 Actor Registry 位置。 | 玩家 B 必须分配至 `game-worker/w3`，绝不能分配至 `game-worker/w2` 或 `game-worker/w1`。这一步验证目标 worker 在迁移收敛前不会进入 agent 候选池。 |
| 6 | 轮询源端迁移状态，直到 gactor 迁移收敛。 | 返回状态显示目标为 `game-worker/w2`，且已迁移 Actor 数量至少为 1；玩家 A 的持续请求无断连、无业务错误。 |

```bash
curl -s http://127.0.0.1:22089/admin/migration \
  -H 'Authorization: Bearer dev-migration-token'
```

| 步骤 | 操作 | 预期结果 |
| --- | --- | --- |
| 7 | 迁移收敛后读取 worker 状态，并再次登录玩家 C。 | `w2=serving`，`migration_id` 已清空。玩家 C 可按 rendezvous 规则被分配至任一 serving worker；若日志选中 `w2`，说明目标已在收敛后恢复接入。 |
| 8 | 对玩家 A 连续发送请求，并检查其 Actor Registry 位置。 | 玩家 A 的请求持续成功，Registry 位置指向 `game-worker/w2`；Player 的登录状态和最后心跳时间被恢复。 |
| 9 | 执行迁移收尾。 | 两个接口均返回 `200 OK`，且源端不再保留已迁移 Player；收尾不能向 agent 下发断开包。 |

```bash
curl -i -X POST http://127.0.0.1:22089/admin/migration/finish-forward \
  -H 'Authorization: Bearer dev-migration-token'

curl -i -X POST http://127.0.0.1:22089/admin/migration/finalize \
  -H 'Authorization: Bearer dev-migration-token'
```

| 步骤 | 操作 | 预期结果 |
| --- | --- | --- |
| 10 | 停止 `w1`，确认其 Redis readiness 记录被删除或 TTL 到期；继续操作玩家 A 和 B。 | `actor:migration:worker:game-worker/w1` 不存在。玩家 A 的客户端连接保持可用，agent 后续请求直接发送至 `w2`，`w1` 不再出现 `HandlePacketRawReq` 或迁移转发日志；玩家 B 继续由 `w3` 服务。 |

## 扩容：一个源迁移至多个目标

本用例独立运行。先恢复 `w1=serving`、`w2=standby`、`w3=serving`、`w4=standby` 的干净测试环境；如运行过前述用例，需重启相关 worker，并确认旧玩家的 Registry 位置和连接不会影响本次测试。准备多个位于 `w1` 的在线玩家，持续发送请求，记录每个 Player ID 和迁移前状态。需要足够多的玩家，以保证 Rendezvous 规划结果同时覆盖 `w2` 和 `w4`，否则补充账号后重新测试。

1. 在 `w1` 发起一对多迁移：

   ```bash
   curl -i -X POST http://127.0.0.1:22089/admin/migration/start \
     -H 'Authorization: Bearer dev-migration-token' \
     -H 'Content-Type: application/json' \
     -d '{"targetNodeIds":["game-worker/w2","game-worker/w4"]}'
   ```

   预期返回 `202 Accepted`。旧的 `targetNodeId` 单目标请求仍受支持，但不能与 `targetNodeIds` 同时使用。

2. 在迁移尚未收敛时检查 Redis：`w1=draining`，`w2=receiving`、`w4=receiving`，源和两个目标的 `migration_id` 相同且非空，`w3=serving`。此时登录新玩家，预期只能选中 `w3`。迁移可能很快收敛，若未捕获 receiving 窗口，不能将这项判为通过；需增加源端在线玩家并重新执行。
3. 轮询 `GET /admin/migration`，直到 `Converged=true`。检查 `Actors` 中各条记录的 `TargetNodeID` 和 `Phase`：已迁移 Player 应分布到 `w2`、`w4`，对应 `Phase=migrated`。混合目标时顶层 `TargetNodeID` 为空，不能用它判断目标。记录 Player ID、目标和迁移统计，核对 Registry 位置与各条目标一致，持续请求无业务错误或断连。
4. 等待收敛回调完成并检查 `w2`、`w4` 均为 `serving` 且 `migration_id` 为空。此时新玩家可被分配到任意 serving worker；多次使用不同账号验证两个新目标均能接入。若某目标仍为 `receiving`，检查源端 `activate migration target` 错误日志，不要继续收尾。
5. 调用前述 `finish-forward`、`finalize` 接口，确认均返回 `200 OK`。停止 `w1` 并等待 readiness 删除或 TTL 到期；继续对分别位于 `w2`、`w4` 的玩家发送请求，确认连接保持、请求直达各自目标，源节点不再承担转发。

### 异常和回滚

- 空目标列表、空节点 ID、重复节点 ID，或同时提供新旧字段：返回 `400`，源和目标状态不变。
- 目标是源节点、非 worker、不在线、非 `standby` 或版本不兼容：返回 `400`，不应预约其他目标或排空源节点。
- 多个源并发预约相同目标：仅一个请求可成功预约；失败请求返回 `409`，它之前预约的其他目标恢复 `standby`，成功请求的目标不受影响。检查各目标的 `instance_id` 和 `migration_id`，避免将其他任务的状态误认作回滚失败。
- 源不是 `serving` 或 gactor 拒绝启动：返回 `409`，本次已预约目标恢复 `standby`；若源已被本次请求切为 `draining`，应恢复 `serving`。Redis 故障可能使回滚失败，此时检查错误日志并人工处理，不应认定回滚成功。

## 验收标准

1. `standby` 和 `receiving` worker 均不会被 agent 用作新 Player 的部署节点。
2. 迁移开始后，源 worker 在迁移期间处于 `draining`，不再接收新 Player。
3. 迁移收敛前，目标 worker 必须保持 `receiving`；收敛后才转换为 `serving`。
4. 已登录玩家 A 的会话、Player 临时状态与请求链路在迁移前后连续可用。
5. `finish-forward` 与 `finalize` 成功后，源 worker 可安全下线，迁移后的 Player 不依赖源节点。
6. 源 worker 下线后，agent 对已迁移 Player 的请求必须直达目标 worker，不能依赖源端转发。
7. 一对多迁移中每个 Actor 的目标在计划冻结后保持不变；所有目标在全局收敛前均不接收新玩家。
8. 任一目标预约失败时，本次此前成功预约的目标应被释放，不得改变其他迁移任务的预约。

## 失败信息收集

失败时保留以下信息，便于定位：

```bash
redis-cli HGETALL actor:migration:worker:game-worker/w1
redis-cli HGETALL actor:migration:worker:game-worker/w2
redis-cli HGETALL actor:migration:worker:game-worker/w3
redis-cli HGETALL actor:migration:worker:game-worker/w4

curl -s http://127.0.0.1:22089/admin/migration \
  -H 'Authorization: Bearer dev-migration-token'
```

同时保存 agent、`w1`、`w2`、`w3`、`w4` 的日志，尤其是 `pick player worker`、`worker migration converged` 和目标预约、激活、释放失败日志。
