package cbctx

import (
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
)

var (
	KeyCallback   = actor.NewCbCtxK[any]("callback")                // 回调上下文键
	KeyPlayer     = actor.NewCbCtxK[*actors.Player]("actor_player") // 玩家上下文键
	KeyChatRoomID = actor.NewCbCtxK[int64]("actor_chat_room_id")    // 聊天室ID上下文键
)
