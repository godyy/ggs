package systems

// System 系统接口.
type System interface {
	// OnStart 启动回调.
	OnStart()

	// OnStop 停止回调.
	OnStop()
}

// systems 系统模块列表.
var systems []System

// RegisterSystem 注册系统模块.
func RegisterSystem[S System](s S) S {
	systems = append(systems, s)
	return s
}

// Start 启动系统模块.
func Start() {
	for _, m := range systems {
		m.OnStart()
	}
}

// cStop 全局停止信号.
var cStop = make(chan struct{}, 1)

// CStop 获取全局停止信号.
func CStop() chan struct{} {
	return cStop
}

// Stop 停止系统模块.
func Stop() {
	close(cStop)
	for _, m := range systems {
		m.OnStop()
	}
}
