package player

import (
	"github.com/godyy/ggs/app/game/internal/handler"
	"github.com/godyy/ggs/internal/infra/actor"
	actorhandler "github.com/godyy/ggs/internal/infra/actor/handler"
	pbc2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
	"google.golang.org/protobuf/proto"
)

func init() {
	initC2SHandler()
	initS2SHandler()
}

func initC2SHandler() {
	registerC2SFunc((*pbc2s.LoginCharacterReq)(nil), false, actorhandler.WrapReqFunc(handleLoginCharacter))
	registerC2SFunc((*pbc2s.HeartbeatReq)(nil), true, actorhandler.WrapReqFunc(handleHeartbeat))
	registerC2SFunc((*pbc2s.ModifyNameReq)(nil), true, actorhandler.WrapReqFunc(handleModifyName))
	registerC2SFunc((*pbc2s.UseItemReq)(nil), true, actorhandler.WrapReqFunc(handleUseItem))
	registerC2SFunc((*pbc2s.SendChatMsgReq)(nil), true, actorhandler.WrapAsyncReqFunc(handleSendChatMsg))
	registerC2SFunc((*pbc2s.CreateGroupChatRoomReq)(nil), true, actorhandler.WrapAsyncReqFunc(handleCreateGroupChatRoom))
	registerS2SFunc((*pbs2s.NotifyJoinChatRoomReq)(nil), actorhandler.WrapRPCNoRespFunc(handleNotifyJoinChatRoom))
	registerC2SFunc((*pbc2s.LeaveChatRoomReq)(nil), true, actorhandler.WrapAsyncReqFunc(handleLeaveChatRoom))
	registerS2SFunc((*pbs2s.NotifyLeaveChatRoomReq)(nil), actorhandler.WrapRPCNoRespFunc(handleNotifyLeaveChatRoom))
	registerC2SFunc((*pbc2s.ChatRoomInviteReq)(nil), true, actorhandler.WrapAsyncReqFunc(handleChatRoomInvite))
	registerC2SFunc((*pbc2s.ChatHistoryReq)(nil), true, actorhandler.WrapAsyncReqFunc(handleChatHistory))
}

func initS2SHandler() {
}

func registerC2SFunc(msg proto.Message, checkLogin bool, f ...actor.HandlerFunc) {
	if checkLogin {
		handler.RegisterC2S(msg, append([]actor.HandlerFunc{mdCheckLogin}, f...)...)
	} else {
		handler.RegisterC2S(msg, f...)
	}
}

func registerS2SFunc(msg proto.Message, f ...actor.HandlerFunc) {
	handler.RegisterS2S(msg, f...)
}
