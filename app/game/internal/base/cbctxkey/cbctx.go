package cbctxkey

import (
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/model/chat"
)

var (
	Callback   = actor.NewCbCtxK[any]("callback")                // 回调上下文键
	Player     = actor.NewCbCtxK[*actors.Player]("actor_player") // 玩家上下文键
	ChatRoomID = actor.NewCbCtxK[int64]("actor_chat_room_id")    // 聊天室ID上下文键
	ChatRoom   = actor.NewCbCtxK[*chat.Room]("chat_room")        // 聊天室
	LastMsgId  = actor.NewCbCtxK[int64]("last_msg_id")           // 最后消息ID
	MemberIds  = actor.NewCbCtxK[[]int64]("member_ids")          // 成员ID列表
	MemberId   = actor.NewCbCtxK[int64]("member_id")             // 成员ID
)
