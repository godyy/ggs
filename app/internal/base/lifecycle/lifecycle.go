package lifecycle

type Callback func()

var (
	beforeStartCallbacks       []Callback
	afterInitDatabaseCallbacks []Callback
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

// RegisterAfterInitDatabase 注册在初始化数据库后调用的回调函数.
func RegisterAfterInitDatabase(cb Callback) {
	afterInitDatabaseCallbacks = append(afterInitDatabaseCallbacks, cb)
}

// AfterInitDatabase 调用初始化数据库后回调函数.
func AfterInitDatabase() {
	for _, cb := range afterInitDatabaseCallbacks {
		cb()
	}
}
