package actors

import (
	"github.com/godyy/ggs/internal/infra/actors/persist"
)

// InitConfig 初始化配置.
type InitConfig struct {
	Persist           *persist.InitConfig // 持久化配置
	DB                string              // 数据库名
	AsyncSaveCallback AsyncSaveCallback   // 异步存储回调
}

var (
	initialized       bool              // 是否初始化
	db                string            // 数据库名
	asyncSaveCallback AsyncSaveCallback // 异步存储回调
)

// Init 初始化.
func Init(cfg *InitConfig) {
	if initialized {
		return
	}
	persist.Init(cfg.Persist)
	db = cfg.DB
	asyncSaveCallback = cfg.AsyncSaveCallback
	initialized = true
}

// checkState 检查状态.
func checkState() {
	if !initialized {
		panic("not initialized")
	}
}
