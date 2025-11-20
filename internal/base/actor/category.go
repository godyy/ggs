package actor

// Actor 分类.
type Category uint16

const (
	CategoryServer Category = 1 // 服务器Actor类别
	CategoryPlayer Category = 2 // 玩家Actor类别
)

func (c Category) Uint16() uint16 {
	return uint16(c)
}

func (c Category) String() string {
	return categoryStrings[c]
}

// categoryStrings 分类字符映射.
var categoryStrings = map[Category]string{
	CategoryServer: "server",
	CategoryPlayer: "player",
}
