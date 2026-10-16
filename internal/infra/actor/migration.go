package actor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggskit/base/db/redis"
	pkgerrors "github.com/pkg/errors"
)

const migrationSnapshotVersion = 1

// WorkerAdmissionState 描述 worker 的玩家接入与迁移状态.
type WorkerAdmissionState string

const (
	WorkerStateStandby   WorkerAdmissionState = "standby"   // 可预约为迁移目标，不接收新玩家。
	WorkerStateReceiving WorkerAdmissionState = "receiving" // 已被迁移任务预约，等待接收迁移 Actor。
	WorkerStateServing   WorkerAdmissionState = "serving"   // 可被 agent 选择并承接新玩家。
	WorkerStateDraining  WorkerAdmissionState = "draining"  // 源 worker 排空中，仅服务既有迁移链路。
)

// WorkerReadiness 描述 worker 存活与接入状态.
type WorkerReadiness struct {
	NodeID          string               `json:"nodeId"`          // Cluster 节点 ID。
	InstanceID      string               `json:"instanceId"`      // 每次进程启动生成，用于隔离同名旧实例。
	BuildVersion    string               `json:"buildVersion"`    // 构建版本，用于迁移兼容性校验。
	ProtocolVersion string               `json:"protocolVersion"` // 协议版本，用于迁移兼容性校验。
	State           WorkerAdmissionState `json:"state"`           // agent 接入与迁移预约状态。
	MigrationID     string               `json:"migrationId"`     // receiving 状态对应的唯一迁移任务 ID。
	UpdatedAt       int64                `json:"updatedAt"`       // 最近一次 worker 存活续期时间。
}

// PlayerMigrationSnapshot 记录迁移期间需要恢复的临时 Player 状态.
type PlayerMigrationSnapshot struct {
	Version           int   `json:"version"`
	ServerID          int64 `json:"serverId"`
	PlayerID          int64 `json:"playerId"`
	IsLogin           bool  `json:"isLogin"`
	LastHeartbeatTime int64 `json:"lastHeartbeatTime"`
	CreatedAt         int64 `json:"createdAt"`
}

// MigrationStore 管理迁移会话快照与 worker 就绪记录.
type MigrationStore struct {
	redisCli    redis.Client
	snapshotTTL time.Duration
}

// NewMigrationStore 创建迁移数据存储.
func NewMigrationStore(redisCli redis.Client, snapshotTTL time.Duration) (*MigrationStore, error) {
	if redisCli == nil {
		return nil, errors.New("migration redis client is required")
	}
	if snapshotTTL <= 0 {
		return nil, errors.New("migration snapshot ttl must > 0")
	}
	return &MigrationStore{redisCli: redisCli, snapshotTTL: snapshotTTL}, nil
}

// workerReadinessKey 返回 worker 状态 Hash 的 Redis 键。
func workerReadinessKey(nodeID string) string {
	return fmt.Sprintf("actor:migration:worker:%s", nodeID)
}

// playerMigrationSnapshotKey 返回指定源租约的 Player 会话快照键。
func playerMigrationSnapshotKey(uid gactor.ActorUID, sourceLeaseID string) string {
	return fmt.Sprintf("actor:migration:%d:%d:%s", uid.Category, uid.ID, sourceLeaseID)
}

// RegisterWorker 注册新 worker，并将接入状态重置为启动时指定的状态。
func (s *MigrationStore) RegisterWorker(ctx context.Context, readiness WorkerReadiness, ttl time.Duration) error {
	if readiness.NodeID == "" || readiness.InstanceID == "" || ttl <= 0 {
		return errors.New("invalid worker readiness")
	}
	if readiness.State != WorkerStateStandby && readiness.State != WorkerStateServing {
		return errors.New("invalid worker initial state")
	}
	readiness.UpdatedAt = time.Now().Unix()
	key := workerReadinessKey(readiness.NodeID)
	if err := s.redisCli.HSet(ctx, key, workerReadinessFields(readiness, true)).Err(); err != nil {
		return err
	}
	return s.redisCli.Expire(ctx, key, ttl).Err()
}

// RefreshWorkerReadiness 续期 worker 存活记录，不覆盖迁移控制设置的状态。
func (s *MigrationStore) RefreshWorkerReadiness(ctx context.Context, readiness WorkerReadiness, ttl time.Duration) error {
	if readiness.NodeID == "" || readiness.InstanceID == "" || ttl <= 0 {
		return errors.New("invalid worker readiness")
	}
	readiness.UpdatedAt = time.Now().Unix()
	key := workerReadinessKey(readiness.NodeID)
	result, err := s.redisCli.Eval(
		ctx,
		refreshWorkerReadinessScript,
		[]string{key},
		readiness.InstanceID,
		readiness.NodeID,
		readiness.BuildVersion,
		readiness.ProtocolVersion,
		readiness.UpdatedAt,
		ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return err
	}
	if result != 1 {
		return errors.New("worker readiness instance changed")
	}
	return nil
}

// workerReadinessFields 将 worker 元数据转换为 Redis Hash 字段。
func workerReadinessFields(readiness WorkerReadiness, withState bool) map[string]any {
	fields := map[string]any{
		"node_id":          readiness.NodeID,
		"instance_id":      readiness.InstanceID,
		"build_version":    readiness.BuildVersion,
		"protocol_version": readiness.ProtocolVersion,
		"updated_at":       readiness.UpdatedAt,
	}
	if withState {
		fields["state"] = string(readiness.State)
		fields["migration_id"] = readiness.MigrationID
	}
	return fields
}

// 续期只能更新存活元数据；状态由迁移控制器通过 CAS 管理，防止目标在 receiving 时被误恢复为 serving。
const refreshWorkerReadinessScript = `
if redis.call("EXISTS", KEYS[1]) == 0 then
	return 0
end
if redis.call("HGET", KEYS[1], "instance_id") ~= ARGV[1] then
	return 0
end
redis.call("HSET", KEYS[1],
        "node_id", ARGV[2],
        "build_version", ARGV[3],
        "protocol_version", ARGV[4],
        "updated_at", ARGV[5])
redis.call("PEXPIRE", KEYS[1], ARGV[6])
return 1
`

// GetWorkerReadiness 获取 worker 就绪记录.
func (s *MigrationStore) GetWorkerReadiness(ctx context.Context, nodeID string) (WorkerReadiness, error) {
	return GetWorkerReadiness(ctx, s.redisCli, nodeID)
}

// GetWorkerReadiness 使用任意 Redis 客户端读取 worker 就绪记录.
func GetWorkerReadiness(ctx context.Context, redisCli redis.Client, nodeID string) (WorkerReadiness, error) {
	if redisCli == nil {
		return WorkerReadiness{}, errors.New("migration redis client is required")
	}
	values, err := redisCli.HGetAll(ctx, workerReadinessKey(nodeID)).Result()
	if err != nil {
		return WorkerReadiness{}, err
	}
	if len(values) == 0 {
		return WorkerReadiness{}, errors.New("worker readiness not found")
	}
	updatedAt, err := strconv.ParseInt(values["updated_at"], 10, 64)
	if err != nil {
		return WorkerReadiness{}, pkgerrors.WithMessage(err, "parse worker updated at")
	}
	readiness := WorkerReadiness{
		NodeID:          values["node_id"],
		InstanceID:      values["instance_id"],
		BuildVersion:    values["build_version"],
		ProtocolVersion: values["protocol_version"],
		State:           WorkerAdmissionState(values["state"]),
		MigrationID:     values["migration_id"],
		UpdatedAt:       updatedAt,
	}
	if readiness.NodeID != nodeID || readiness.InstanceID == "" {
		return WorkerReadiness{}, errors.New("invalid worker readiness")
	}
	return readiness, nil
}

// 状态转换同时校验实例、旧状态和迁移任务，避免并发迁移或同名 worker 重启覆盖新状态。
const transitionWorkerStateScript = `
if redis.call("EXISTS", KEYS[1]) == 0 then
	return 0
end
if redis.call("HGET", KEYS[1], "instance_id") ~= ARGV[1] then
	return 0
end
if redis.call("HGET", KEYS[1], "state") ~= ARGV[2] then
	return 0
end
if redis.call("HGET", KEYS[1], "migration_id") ~= ARGV[3] then
	return 0
end
redis.call("HSET", KEYS[1], "state", ARGV[4], "migration_id", ARGV[5])
return 1
`

// TransitionWorkerState 以实例与当前状态为条件原子变更 worker 接入状态。
func (s *MigrationStore) TransitionWorkerState(
	ctx context.Context,
	nodeID, instanceID string,
	from WorkerAdmissionState,
	fromMigrationID string,
	to WorkerAdmissionState,
	toMigrationID string,
) (bool, error) {
	if nodeID == "" || instanceID == "" {
		return false, errors.New("worker node id and instance id are required")
	}
	result, err := s.redisCli.Eval(
		ctx,
		transitionWorkerStateScript,
		[]string{workerReadinessKey(nodeID)},
		instanceID,
		string(from),
		fromMigrationID,
		string(to),
		toMigrationID,
	).Int64()
	if err != nil {
		return false, pkgerrors.WithMessage(err, "transition worker state")
	}
	return result == 1, nil
}

// 删除时校验实例 ID，避免旧进程退出时清理同名新 worker 的记录。
const deleteWorkerReadinessScript = `
if redis.call("HGET", KEYS[1], "instance_id") ~= ARGV[1] then
	return 0
end
return redis.call("DEL", KEYS[1])
`

// DeleteWorkerReadiness 删除当前实例的 worker 就绪记录.
func (s *MigrationStore) DeleteWorkerReadiness(ctx context.Context, nodeID, instanceID string) error {
	if nodeID == "" || instanceID == "" {
		return errors.New("worker node id and instance id are required")
	}
	_, err := s.redisCli.Eval(
		ctx,
		deleteWorkerReadinessScript,
		[]string{workerReadinessKey(nodeID)},
		instanceID,
	).Int64()
	return err
}

// SavePlayerSnapshot 保存可幂等读取的 Player 会话快照.
func (s *MigrationStore) SavePlayerSnapshot(ctx context.Context, uid gactor.ActorUID, sourceLeaseID string, snapshot PlayerMigrationSnapshot) error {
	if sourceLeaseID == "" {
		return errors.New("source lease id is required")
	}
	snapshot.Version = migrationSnapshotVersion
	snapshot.PlayerID = uid.ID
	snapshot.CreatedAt = time.Now().Unix()
	value, err := json.Marshal(snapshot)
	if err != nil {
		return pkgerrors.WithMessage(err, "marshal player migration snapshot")
	}
	return s.redisCli.Set(ctx, playerMigrationSnapshotKey(uid, sourceLeaseID), value, s.snapshotTTL).Err()
}

// LoadPlayerSnapshot 获取 Player 会话快照.
func (s *MigrationStore) LoadPlayerSnapshot(ctx context.Context, uid gactor.ActorUID, sourceLeaseID string) (PlayerMigrationSnapshot, error) {
	value, err := s.redisCli.Get(ctx, playerMigrationSnapshotKey(uid, sourceLeaseID)).Bytes()
	if err != nil {
		return PlayerMigrationSnapshot{}, err
	}
	var snapshot PlayerMigrationSnapshot
	if err := json.Unmarshal(value, &snapshot); err != nil {
		return PlayerMigrationSnapshot{}, pkgerrors.WithMessage(err, "unmarshal player migration snapshot")
	}
	if snapshot.Version != migrationSnapshotVersion || snapshot.PlayerID != uid.ID {
		return PlayerMigrationSnapshot{}, errors.New("invalid player migration snapshot")
	}
	return snapshot, nil
}

// DeletePlayerSnapshot 删除指定迁移快照.
func (s *MigrationStore) DeletePlayerSnapshot(ctx context.Context, uid gactor.ActorUID, sourceLeaseID string) error {
	return s.redisCli.Del(ctx, playerMigrationSnapshotKey(uid, sourceLeaseID)).Err()
}
