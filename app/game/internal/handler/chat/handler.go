package chat

import (
	"github.com/godyy/ggs/app/game/internal/handler"
	actorhandler "github.com/godyy/ggs/internal/infra/actor/handler"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
)

func init() {
	handler.RegisterS2S((*pbs2s.SendChatMsgReq)(nil), actorhandler.WrapAsyncRPCFunc(handleSendChatMsg))
	handler.RegisterS2S((*pbs2s.CreateGroupChatRoomReq)(nil), actorhandler.WrapAsyncRPCFunc(handleCreateGroupChatRoom))
	handler.RegisterS2S((*pbs2s.LeaveChatRoomReq)(nil), actorhandler.WrapAsyncRPCFunc(handleLeaveChatRoom))
	handler.RegisterS2S((*pbs2s.ChatRoomInviteReq)(nil), actorhandler.WrapAsyncRPCFunc(handleChatRoomInvite))
	handler.RegisterS2S((*pbs2s.ChatHistoryReq)(nil), actorhandler.WrapAsyncRPCFunc(handleChatHistory))
}
