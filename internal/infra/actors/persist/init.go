package persist

// InitConfig 初始化配置.
type InitConfig struct {
	BD BD // 持久化后台
}

var (
	initialized bool // 是否已初始化
	bd          BD   // 持久化后台
)

// Init 初始化.
func Init(cfg *InitConfig) {
	if initialized {
		return
	}
	bd = cfg.BD
	initialized = true
}

// checkState 检查状态.
func checkState() {
	if !initialized {
		panic("not initialized")
	}
}
