package player

// BaseInfo 基础信息
type BaseInfo struct {
	moduleBase[*BaseInfo]
	Name string `bson:"name"` // 昵称
}

func (b *BaseInfo) Key() string {
	return "base"
}
