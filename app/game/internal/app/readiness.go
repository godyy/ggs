package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	pkgerrors "github.com/pkg/errors"
)

// startWorkerReadiness 注册 worker 状态并启动 TTL 续期。
func (a *app) startWorkerReadiness() error {
	if !a.env.IsWorker() {
		return nil
	}
	if a.config.Migration.ReadinessTTL <= 0 {
		return errors.New("migration readiness ttl must > 0")
	}
	if a.config.Migration.BuildVersion == "" || a.config.Migration.ProtocolVersion == "" {
		return errors.New("migration build and protocol versions are required")
	}
	instanceID, err := newWorkerInstanceID()
	if err != nil {
		return err
	}
	a.workerInstanceID = instanceID
	// 注册会重置同名旧实例遗留的状态；后续续期不会覆盖迁移控制设置的状态。
	if err := a.migrationStore.RegisterWorker(context.Background(), a.workerReadiness(), a.config.Migration.ReadinessTTL); err != nil {
		return pkgerrors.WithMessage(err, "register worker readiness")
	}

	a.readinessStop = make(chan struct{})
	go func() {
		ticker := time.NewTicker(a.config.Migration.ReadinessTTL / 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := a.refreshWorkerReadiness(); err != nil {
					// TTL 到期会自动将 worker 从候选集中移除。
					logger.Get().Errorf("refresh worker readiness failed: %v", err)
				}
			case <-a.readinessStop:
				return
			}
		}
	}()
	return nil
}

// stopWorkerReadiness 停止续期并删除当前实例的 worker 状态。
func (a *app) stopWorkerReadiness() {
	if !a.env.IsWorker() || a.migrationStore == nil {
		return
	}
	if a.readinessStop != nil {
		close(a.readinessStop)
		a.readinessStop = nil
	}
	if err := a.migrationStore.DeleteWorkerReadiness(context.Background(), selfNode("").GetNodeId(), a.workerInstanceID); err != nil {
		logger.Get().Errorf("delete worker readiness failed: %v", err)
	}
}

// refreshWorkerReadiness 仅续期当前实例的存活元数据。
func (a *app) refreshWorkerReadiness() error {
	return a.migrationStore.RefreshWorkerReadiness(context.Background(), a.workerReadiness(), a.config.Migration.ReadinessTTL)
}

// workerReadiness 构造当前 worker 的状态记录。
func (a *app) workerReadiness() actor.WorkerReadiness {
	return actor.WorkerReadiness{
		NodeID:          selfNode("").GetNodeId(),
		InstanceID:      a.workerInstanceID,
		BuildVersion:    a.config.Migration.BuildVersion,
		ProtocolVersion: a.config.Migration.ProtocolVersion,
		// 仅在初始注册时生效；续期路径不写入该字段。
		State: actor.WorkerAdmissionState(a.env.WorkerState()),
	}
}

// newWorkerInstanceID 生成进程级唯一标识，用于隔离同名 worker 实例。
func newWorkerInstanceID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", pkgerrors.WithMessage(err, "generate worker instance id")
	}
	return hex.EncodeToString(value[:]), nil
}
