package env

const (
	StageDev  = "dev"  // 开发环境
	StageProd = "prod" // 生产环境
)

// Accessor 环境变量访问器.
type Accessor struct{}

var (
	ac = &Accessor{}

	stage string = StageDev // 环境 dev/prod

	debug bool // 是否启用调试模式
)

// All 返回所有环境变量.
func All() *Accessor {
	return ac
}

// Stage 返回当前环境.
func (*Accessor) Stage() string {
	return stage
}

// Dev 返回是否为开发环境.
func (*Accessor) Dev() bool {
	return stage == StageDev
}

// Prod 返回是否为生产环境.
func (*Accessor) Prod() bool {
	return stage == StageProd
}

// Debug 返回是否启用调试模式.
func (a *Accessor) Debug() bool {
	return debug && !a.Prod()
}
