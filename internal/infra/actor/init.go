package actor

import (
	"errors"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actor/persist"
	"github.com/godyy/ggskit/infra/actor"
)

// InitConfig 初始化配置.
type InitConfig struct {
	Persist           *persist.InitConfig // 持久化配置
	DB                string              // 数据库名
	AsyncSaveCallback AsyncSaveCallback   // 异步存储回调
	ActorSaveDelay    time.Duration       // actor 保存延迟
	MigrationStore    *MigrationStore     // 迁移快照存储
	ServerStore       *ServerStore        // Player 稳定区服归属存储
}

var (
	initialized       bool              // 是否初始化
	db                string            // 数据库名
	asyncSaveCallback AsyncSaveCallback // 异步存储回调
	actorSaveDelay    time.Duration     // actor 保存延迟
	migrationStore    *MigrationStore   // 迁移快照存储
	serverStore       *ServerStore      // Player 稳定区服归属存储
)

// Init 初始化.
func Init(cfg *InitConfig) {
	if initialized {
		return
	}
	persist.Init(cfg.Persist)
	db = cfg.DB
	asyncSaveCallback = cfg.AsyncSaveCallback
	actorSaveDelay = cfg.ActorSaveDelay
	migrationStore = cfg.MigrationStore
	serverStore = cfg.ServerStore
	actorSugarUtil = actor.NewActorSugarUtil(ProtoRegistry)
	contextSugarUtil = actor.NewContextSugarUtil(ProtoRegistry)
	initialized = true
}

// GetMigrationStore 获取迁移快照存储.
func GetMigrationStore() *MigrationStore {
	return migrationStore
}

// GetActorServerID 获取 Actor 的稳定区服归属。
func GetActorServerID(uid gactor.ActorUID) (int64, bool, error) {
	if serverStore == nil {
		return 0, false, errors.New("actor server store is not initialized")
	}
	return serverStore.GetActorServer(uid)
}

// checkState 检查状态.
func checkState() {
	if !initialized {
		panic("not initialized")
	}
}
