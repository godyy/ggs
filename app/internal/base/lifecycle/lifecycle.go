package lifecycle

type Callback func()

var (
	beforeStartCallbacks []Callback
)

// RegisterBeforeStart 注册在启动前调用的回调函数.
func RegisterBeforeStart(cb Callback) {
	beforeStartCallbacks = append(beforeStartCallbacks, cb)
}

// BeforeStart 调用启动前回调函数.
func BeforeStart() {
	for _, cb := range beforeStartCallbacks {
		cb()
	}
}
