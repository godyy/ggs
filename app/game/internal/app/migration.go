package app

import (
	"context"
	"crypto/subtle"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggskit/infra/noderouter"
	pkgerrors "github.com/pkg/errors"
)

type migrationStartRequest struct {
	TargetNodeIDs []string `json:"targetNodeIds"`
}

// targets 保留旧的单目标请求格式，并拒绝空值和重复目标。
func (r migrationStartRequest) targets() ([]string, error) {
	if len(r.TargetNodeIDs) == 0 {
		return nil, errors.New("migration targets are required")
	}
	seen := make(map[string]struct{}, len(r.TargetNodeIDs))
	for _, id := range r.TargetNodeIDs {
		if id == "" {
			return nil, errors.New("migration target is empty")
		}
		if _, ok := seen[id]; ok {
			return nil, errors.New("duplicate migration target")
		}
		seen[id] = struct{}{}
	}
	return r.TargetNodeIDs, nil
}

// migrationPlanner 使用与 agent 选点一致的 Player ID 和 Rendezvous 哈希。
func migrationPlanner(targets []string) gactor.ServiceMigrationPlanner {
	selector := noderouter.NewRendezvousSelector()
	selector.Set(map[string][]string{consts.NodeGameWorker: targets}, true)
	return func(uid gactor.ActorUID) (string, error) {
		var key [8]byte
		binary.NativeEndian.PutUint64(key[:], uint64(uid.ID))
		selected := selector.Pick(consts.NodeGameWorker, key[:], 1)
		if len(selected) == 0 {
			return "", errors.New("no migration target available")
		}
		return selected[0], nil
	}
}

// reserveMigrationTargets 逐一预约目标，失败时释放此前已预约的目标。
func reserveMigrationTargets(ctx context.Context, targets []actor.WorkerReadiness, migrationID string,
	reserve func(context.Context, actor.WorkerReadiness, string) error,
	release func(context.Context, actor.WorkerReadiness, string),
) error {
	for i, target := range targets {
		if err := reserve(ctx, target, migrationID); err != nil {
			// 超时可能发生在 Redis 已执行之后，也要尝试释放当前目标。
			for _, reserved := range targets[:i+1] {
				release(context.Background(), reserved, migrationID)
			}
			return pkgerrors.WithMessagef(err, "reserve migration target %s", target.NodeID)
		}
	}
	return nil
}

// registerMigrationHTTP 注册仅 worker 可用的迁移管理接口。
func (a *app) registerMigrationHTTP(mux *http.ServeMux) {
	if !a.env.IsWorker() {
		return
	}
	mux.HandleFunc("POST /admin/migration/start", a.handleMigrationStart)
	mux.HandleFunc("GET /admin/migration", a.handleMigrationStatus)
	mux.HandleFunc("POST /admin/migration/finish-forward", a.handleMigrationFinishForward)
	mux.HandleFunc("POST /admin/migration/finalize", a.handleMigrationFinalize)
}

// checkMigrationAuth 校验迁移管理接口的 Bearer Token。
func (a *app) checkMigrationAuth(w http.ResponseWriter, r *http.Request) bool {
	expected := a.config.Migration.AdminToken
	if expected == "" {
		http.Error(w, "migration admin token is not configured", http.StatusServiceUnavailable)
		return false
	}
	const bearer = "Bearer "
	auth := r.Header.Get("Authorization")
	if len(auth) < len(bearer) || auth[:len(bearer)] != bearer ||
		subtle.ConstantTimeCompare([]byte(auth[len(bearer):]), []byte(expected)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

// handleMigrationStart 预约目标、排空源 worker 并启动 gactor 全量迁移。
func (a *app) handleMigrationStart(w http.ResponseWriter, r *http.Request) {
	if !a.checkMigrationAuth(w, r) {
		return
	}
	var request migrationStartRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	targetIDs, err := request.targets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	targets := make([]actor.WorkerReadiness, 0, len(targetIDs))
	for _, id := range targetIDs {
		target, err := a.validateMigrationTarget(r.Context(), id)
		if err != nil {
			http.Error(w, pkgerrors.WithMessagef(err, "validate migration target %s", id).Error(), http.StatusBadRequest)
			return
		}
		targets = append(targets, target)
	}

	migrationID, err := newWorkerInstanceID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 先独占预约目标，避免多个源 worker 同时将 Actor 迁入同一节点。
	if err := reserveMigrationTargets(r.Context(), targets, migrationID, a.reserveMigrationTarget, a.releaseMigrationTarget); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	releaseTargets := func() {
		for _, target := range targets {
			a.releaseMigrationTarget(context.Background(), target, migrationID)
		}
	}
	// 源进入 draining 后立即从 agent 候选中移除，迁移失败时再回滚至 serving。
	if err := a.transitionSelfWorkerState(r.Context(), actor.WorkerStateServing, "", actor.WorkerStateDraining, migrationID); err != nil {
		// 即使调用超时，Redis 也可能已经完成转换；仅回滚本次任务的状态。
		if rollbackErr := a.transitionSelfWorkerState(context.Background(), actor.WorkerStateDraining, migrationID, actor.WorkerStateServing, ""); rollbackErr != nil {
			logger.Get().Warnf("restore source worker state after failed transition: %v", rollbackErr)
		}
		releaseTargets()
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err := a.actorService.StartMigrationWithPlanner(migrationPlanner(targetIDs), func(snapshot gactor.ServiceMigrationSnapshot) {
		logger.Get().Infof(
			"worker migration converged: targets=%v actors=%d migrated=%d retries=%d",
			targetIDs,
			snapshot.TrackedActorCount,
			snapshot.MigratedActorCount,
			snapshot.RetryCount,
		)
		// 仅在 gactor 已收敛后开放目标接入，防止迁移中的新 Player 混入目标。
		for _, target := range targets {
			if err := a.activateMigrationTarget(context.Background(), target, migrationID); err != nil {
				logger.Get().Errorf("activate migration target %s failed: %v", target.NodeID, err)
			}
		}
	}); err != nil {
		// StartMigration 未建立迁移上下文时，恢复两端的接入状态。
		if rollbackErr := a.transitionSelfWorkerState(context.Background(), actor.WorkerStateDraining, migrationID, actor.WorkerStateServing, ""); rollbackErr != nil {
			logger.Get().Errorf("restore source worker state failed: %v", rollbackErr)
		}
		releaseTargets()
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeMigrationJSON(w, http.StatusAccepted, a.actorService.MigrationSnapshot())
}

// handleMigrationStatus 返回当前 worker 的 gactor 迁移快照。
func (a *app) handleMigrationStatus(w http.ResponseWriter, r *http.Request) {
	if !a.checkMigrationAuth(w, r) {
		return
	}
	writeMigrationJSON(w, http.StatusOK, a.actorService.MigrationSnapshot())
}

// handleMigrationFinishForward 在迁移收敛后关闭源端转发。
func (a *app) handleMigrationFinishForward(w http.ResponseWriter, r *http.Request) {
	if !a.checkMigrationAuth(w, r) {
		return
	}
	if err := a.actorService.FinishMigrationForwarding(); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeMigrationJSON(w, http.StatusOK, a.actorService.MigrationSnapshot())
}

// handleMigrationFinalize 回收源 worker 中已迁移的 Actor。
func (a *app) handleMigrationFinalize(w http.ResponseWriter, r *http.Request) {
	if !a.checkMigrationAuth(w, r) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err := a.actorService.FinalizeMigration(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeMigrationJSON(w, http.StatusOK, a.actorService.MigrationSnapshot())
}

// validateMigrationTarget 校验全局 worker 目标在线、兼容且仍处于 standby 状态。
func (a *app) validateMigrationTarget(ctx context.Context, targetNodeID string) (actor.WorkerReadiness, error) {
	if targetNodeID == "" || targetNodeID == selfNode("").GetNodeId() {
		return actor.WorkerReadiness{}, errors.New("migration target is invalid")
	}
	node, err := a.cluster.GetNode(targetNodeID)
	if err != nil {
		return actor.WorkerReadiness{}, pkgerrors.WithMessage(err, "get migration target node")
	}
	if node.Category != consts.NodeGameWorker {
		return actor.WorkerReadiness{}, errors.New("migration target must be a worker")
	}
	readiness, err := actor.GetWorkerReadiness(ctx, a.redisClient, targetNodeID)
	if err != nil {
		return actor.WorkerReadiness{}, pkgerrors.WithMessage(err, "get migration target readiness")
	}
	if readiness.State != actor.WorkerStateStandby || readiness.NodeID != targetNodeID ||
		readiness.BuildVersion != a.config.Migration.BuildVersion ||
		readiness.ProtocolVersion != a.config.Migration.ProtocolVersion {
		return actor.WorkerReadiness{}, errors.New("migration target is not standby or version compatible")
	}
	return readiness, nil
}

// reserveMigrationTarget 原子预约 standby 目标，使其进入 receiving 状态。
func (a *app) reserveMigrationTarget(ctx context.Context, target actor.WorkerReadiness, migrationID string) error {
	// standby -> receiving：目标保持在线但不会被 agent 选择。
	ok, err := a.migrationStore.TransitionWorkerState(
		ctx,
		target.NodeID,
		target.InstanceID,
		actor.WorkerStateStandby,
		"",
		actor.WorkerStateReceiving,
		migrationID,
	)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("migration target state changed")
	}
	return nil
}

// activateMigrationTarget 在迁移收敛后原子开放目标接入。
func (a *app) activateMigrationTarget(ctx context.Context, target actor.WorkerReadiness, migrationID string) error {
	// receiving -> serving：仅当前迁移任务可在收敛后开放目标接入。
	ok, err := a.migrationStore.TransitionWorkerState(
		ctx,
		target.NodeID,
		target.InstanceID,
		actor.WorkerStateReceiving,
		migrationID,
		actor.WorkerStateServing,
		"",
	)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("migration target activation state changed")
	}
	return nil
}

// releaseMigrationTarget 释放本次失败迁移预约的目标。
func (a *app) releaseMigrationTarget(ctx context.Context, target actor.WorkerReadiness, migrationID string) {
	// 失败回滚只释放本次预约，避免误释放被其他流程接管后的目标。
	if ok, err := a.migrationStore.TransitionWorkerState(
		ctx,
		target.NodeID,
		target.InstanceID,
		actor.WorkerStateReceiving,
		migrationID,
		actor.WorkerStateStandby,
		"",
	); err != nil {
		logger.Get().Errorf("release migration target %s failed: %v", target.NodeID, err)
	} else if !ok {
		logger.Get().Warnf("release migration target %s ignored because state changed", target.NodeID)
	}
}

// transitionSelfWorkerState 原子更新源 worker 的接入状态。
func (a *app) transitionSelfWorkerState(ctx context.Context, from actor.WorkerAdmissionState, fromMigrationID string, to actor.WorkerAdmissionState, toMigrationID string) error {
	// 用任务 ID 区分并发迁移，避免失败请求回滚其他任务的源状态。
	ok, err := a.migrationStore.TransitionWorkerState(
		ctx,
		selfNode("").GetNodeId(),
		a.workerInstanceID,
		from,
		fromMigrationID,
		to,
		toMigrationID,
	)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("source worker state changed")
	}
	return nil
}

// writeMigrationJSON 输出迁移管理接口的 JSON 响应。
func writeMigrationJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
