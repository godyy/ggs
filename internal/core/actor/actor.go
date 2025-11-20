package actor

// Actor 分类.
const (
	CategoryServer = uint16(1) // 服务器Actor类别
	CategoryPlayer = uint16(2) // 玩家Actor类别
)

// categoryNames 分类名称映射.
var categoryNames = map[uint16]string{
	CategoryServer: "server",
	CategoryPlayer: "player",
}

// CategoryName 获取分类名称.
func CategoryName(category uint16) string {
	return categoryNames[category]
}
