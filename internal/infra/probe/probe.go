package probe

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// ErrInitialized 已初始化错误.
var ErrInitialized = errors.New("probe already initialized")

// ErrNotInitialized 未初始化错误.
var ErrNotInitialized = errors.New("probe not initialized")

var (
	initialized bool      // 是否已初始化
	startAt     time.Time // 启动时间
)

// Init 初始化
func Init(opts ...Option) {
	checkInitialized(false)
	for _, opt := range opts {
		opt()
	}
	startAt = time.Now()
	initialized = true
}

// HealthzHandler 原生 http 健康探针.
func HealthzHandler() http.HandlerFunc {
	checkInitialized(true)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":         "ok",
			"uptime_seconds": int(time.Since(startAt).Seconds()),
		})
	}
}

// checkInitialized 检查初始化状态.
func checkInitialized(b bool) {
	if b && !initialized {
		panic(ErrNotInitialized)
	}
	if !b && initialized {
		panic(ErrInitialized)
	}
}
