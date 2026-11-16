package chat

import (
	"github.com/godyy/ggs/app/game/internal/base/cbctxkey"
	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/handler"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
)

// handleSendChatMsg 处理发送聊天消息.
func handleSendChatMsg(ctx *actor.Context, req *pbs2s.SendChatMsgReq) error {
	cs := actor.NewContextSuspender(ctx).Suspend()
	systems.Chat.SendMsg(actor.CtxActor[*actors.ChatMgr](ctx), systems.SendMsgParams{
		RoomId:   req.RoomId,
		SenderId: req.SenderId,
		Content:  req.Content,
	}, func(mgr *actors.ChatMgr, err error) {
		defer cs.Resume()
		if err != nil {
			handler.ReplyErrorAbort(cs.Context, err)
			return
		}
		handler.ReplySuccess(cs.Context)
	})
	return nil
}

// handleCreateGroupChatRoom 处理创建群聊室请求.
func handleCreateGroupChatRoom(ctx *actor.Context, req *pbs2s.CreateGroupChatRoomReq) error {
	cs := actor.NewContextSuspender(ctx).Suspend()
	systems.Chat.CreateGroupChatRoom(actor.CtxActor[*actors.ChatMgr](ctx), &systems.CreateGroupChatRoomParams{
		Name:      req.Name,
		MemberIds: req.MemberIds,
	}, func(mgr *actors.ChatMgr, err error) {
		defer cs.Resume()
		if err != nil {
			handler.ReplyErrorAbort(cs.Context, err)
			return
		}
		handler.ReplySuccess(cs.Context)
	})
	return nil
}

// handleLeaveChatRoom 处理退出聊天室请求.
func handleLeaveChatRoom(ctx *actor.Context, req *pbs2s.LeaveChatRoomReq) error {
	cs := actor.NewContextSuspender(ctx).Suspend()
	systems.Chat.LeaveChatRoom(actor.CtxActor[*actors.ChatMgr](ctx), req.RoomId, req.MemberId, func(mgr *actors.ChatMgr, err error) {
		defer cs.Resume()
		if err != nil {
			handler.ReplyErrorAbort(cs.Context, err)
			return
		}
		handler.ReplySuccess(cs.Context)
	})
	return nil
}

// handleChatRoomInvite 处理聊天室邀请请求.
func handleChatRoomInvite(ctx *actor.Context, req *pbs2s.ChatRoomInviteReq) error {
	cs := actor.NewContextSuspender(ctx).Suspend()
	systems.Chat.ChatRoomInvite(actor.CtxActor[*actors.ChatMgr](ctx), &systems.ChatRoomInviteParams{
		RoomId:    req.RoomId,
		InviterId: req.InviterId,
		TargetIds: req.TargetIds,
	}, func(mgr *actors.ChatMgr, err error) {
		defer cs.Resume()
		if err != nil {
			handler.ReplyErrorAbort(cs.Context, err)
			return
		}
		handler.ReplySuccess(cs.Context)
	})
	return nil
}

// handleChatHistory 处理获取聊天历史请求.
func handleChatHistory(ctx *actor.Context, req *pbs2s.ChatHistoryReq) error {
	systems.Chat.LoadRoomAsync(ctx, req.RoomId, onLoadRoom2GetHistory, actor.NewCbCtx(
		cbctxkey.MemberId, req.MemberId,
		cbctxkey.LastMsgId, req.LastMsgId,
		cbctxkey.ChatRoomID, req.RoomId,
	))
	return nil
}

// onLoadRoom2GetHistory 处理加载聊天室后获取聊天历史.
func onLoadRoom2GetHistory(resp systems.ChatRoomLoadResp) {
	cctx := resp.Op.(*actor.Context)
	cc := resp.CbCtx.(*actor.CbCtx)
	defer actor.CbCtxRelease(cc)
	memberId := actor.CbCtxMust(cc, cbctxkey.MemberId)
	roomId := actor.CbCtxMust(cc, cbctxkey.ChatRoomID)
	lastMsgId := actor.CbCtxMust(cc, cbctxkey.LastMsgId)

	if resp.Err != nil {
		logger.Get().Errorf("[handleChatHistory] player %d, load room %d failed, err: %v", memberId, roomId, resp.Err)
		return
	}
	if err := systems.Chat.CheckChatHistoryPriviledge(resp.Room, memberId); err != nil {
		logger.Get().Errorf("[handleChatHistory] player %d, check chat history priviledge failed, err: %v", memberId, roomId, err)
		handler.ReplyErrorAbort(cctx, err)
		return
	}

	handler.Reply(cctx, &pbs2s.ChatHistoryResp{
		RoomId: resp.Room.ID,
		Msgs:   systems.Chat.PackPBHistorys(systems.Chat.GetRoomHistory(resp.Room, lastMsgId)),
	})
}
