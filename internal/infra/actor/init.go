package actor

import (
	"time"

	"github.com/godyy/ggs/internal/infra/actor/persist"
	protoreg "github.com/godyy/ggs/internal/protocol/registry"
	"github.com/godyy/ggskit/infra/actor"
)

// InitConfig 初始化配置.
type InitConfig struct {
	Persist           *persist.InitConfig // 持久化配置
	DB                string              // 数据库名
	AsyncSaveCallback AsyncSaveCallback   // 异步存储回调
	ActorSaveDelay    time.Duration       // actor 保存延迟
}

var (
	initialized       bool              // 是否初始化
	db                string            // 数据库名
	asyncSaveCallback AsyncSaveCallback // 异步存储回调
	actorSaveDelay    time.Duration     // actor 保存延迟
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
	actorSugarUtil = actor.NewActorSugarUtil(protoreg.Registry)
	contextSugarUtil = actor.NewContextSugarUtil(protoreg.Registry)
	initialized = true
}

// checkState 检查状态.
func checkState() {
	if !initialized {
		panic("not initialized")
	}
}
