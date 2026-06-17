package player

// BaseInfo 基础信息
type BaseInfo struct {
	Name string `bson:"name"` // 昵称
}

func (b *BaseInfo) OnInit() {}

func (b *BaseInfo) ModuleKey() string {
	return "base"
}
