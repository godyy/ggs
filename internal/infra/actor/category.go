package actor

import "github.com/godyy/ggskit/infra/actor"

// Category 分类.
type Category actor.Category

const (
	CategoryServer  Category = 1 // 服务器Actor类别
	CategoryPlayer  Category = 2 // 玩家Actor类别
	CategoryChatMgr Category = 3 // 聊天管理Actor类别
)

func (c Category) ActorCategory() actor.Category {
	return actor.Category(c)
}

func (c Category) String() string {
	return categoryStrings[c]
}

// categoryStrings 分类字符映射.
var categoryStrings = map[Category]string{
	CategoryServer:  "server",
	CategoryPlayer:  "player",
	CategoryChatMgr: "chat_mgr",
}

func PlayerUID(id int64) ActorUID {
	return ActorUID{
		Category: CategoryPlayer.ActorCategory(),
		ID:       id,
	}
}
