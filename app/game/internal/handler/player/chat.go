package player

import (
	"time"

	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/gdconf"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/handler"
	pbc2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/infra/actor/protocol/pb/common"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
)

// handleSendChatMsg 处理发送聊天消息请求.
func handleSendChatMsg(ctx *actor.Context, req *pbc2s.SendChatMsgReq) error {
	cs := actor.NewContextSuspender(ctx).Suspend()
	systems.Chat.PlayerSendMsg(actor.CtxActor[*actors.Player](ctx), req.RoomId, req.Content, func(p *actors.Player, err error) {
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
func handleCreateGroupChatRoom(ctx *actor.Context, req *pbc2s.CreateGroupChatRoomReq) error {
	cs := actor.NewContextSuspender(ctx).Suspend()
	systems.Chat.PlayerCreateRoom(actor.CtxActor[*actors.Player](ctx), systems.PlayerCreateRoomParams{
		RoomType:  int32(gdconf.ChatRoomTypeGroup),
		RoomName:  req.Name,
		MemberIds: req.MemberIds,
	}, func(p *actors.Player, err error) {
		defer cs.Resume()
		if err != nil {
			handler.ReplyErrorAbort(cs.Context, err)
			return
		}
		handler.ReplySuccess(cs.Context)
	})
	return nil
}

// handleNotifyJoinChatRoom 处理通知加入聊天室请求.
func handleNotifyJoinChatRoom(ctx *actor.Context, req *pbs2s.NotifyJoinChatRoomReq) error {
	return systems.Chat.PlayerAddRoom(actor.CtxActor[*actors.Player](ctx), req.RoomId, req.RoomType)
}

// handleLeaveChatRoom 处理退出聊天室请求.
func handleLeaveChatRoom(ctx *actor.Context, req *pbc2s.LeaveChatRoomReq) error {
	cs := actor.NewContextSuspender(ctx).Suspend()
	systems.Chat.PlayerLeaveChatRoom(actor.CtxActor[*actors.Player](ctx), req.RoomId, func(p *actors.Player, err error) {
		defer cs.Resume()
		if err != nil {
			handler.ReplyErrorAbort(cs.Context, err)
			return
		}
		handler.ReplySuccess(cs.Context)
	})
	return nil
}

// handleNotifyLeaveChatRoom 处理通知退出聊天室请求.
func handleNotifyLeaveChatRoom(ctx *actor.Context, req *pbs2s.NotifyLeaveChatRoomReq) error {
	return systems.Chat.PlayerDelRoom(actor.CtxActor[*actors.Player](ctx), req.RoomId)
}

// handleChatRoomInvite 处理邀请玩家加入聊天室请求.
func handleChatRoomInvite(ctx *actor.Context, req *pbc2s.ChatRoomInviteReq) error {
	cs := actor.NewContextSuspender(ctx).Suspend()
	systems.Chat.PlayerChatRoomInvite(actor.CtxActor[*actors.Player](ctx), req.RoomId, req.TargetIds, func(p *actors.Player, err error) {
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
func handleChatHistory(ctx *actor.Context, req *pbc2s.ChatHistoryReq) error {
	p := actor.CtxActor[*actors.Player](ctx)
	if err := actor.SugarContext(ctx).AsyncRPCWithTimeout(systems.Chat.MgrUID(), &pbs2s.ChatHistoryReq{
		MemberId:  p.ID(),
		RoomId:    req.RoomId,
		LastMsgId: req.LastMsgId,
	}, func(resp actor.ContextRPCResp) {
		result := actor.NewRPCResult(resp.Reply, resp.Err)
		if !result.Success() {
			handler.ReplyErrorAbort(ctx, result.Err())
			return
		}
		historyResp := result.Reply().(*pbs2s.ChatHistoryResp)
		handler.Reply(ctx, &pbc2s.ChatHistoryResp{
			RoomId: historyResp.RoomId,
			Msgs:   historyResp.Msgs,
		})
	}, 5*time.Second); err != nil {
		logger.Get().Errorf("[handleChatHistory] player %d async get room %d history failed, err: %v", p.ID(), req.RoomId, err)
		return actor.WithPbError(pbcommon.ErrCode_ECInternalError)
	}
	return nil
}
