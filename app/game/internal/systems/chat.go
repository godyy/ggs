package systems

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/app"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/gdconf"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/lifecycle"
	"github.com/godyy/ggs/internal/infra/actor/model"
	"github.com/godyy/ggs/internal/infra/actor/model/chat"
	"github.com/godyy/ggs/internal/infra/actor/model/player"
	pbc2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/infra/actor/protocol/pb/common"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
	"github.com/godyy/ggs/internal/infra/systems"
	"github.com/godyy/ggskit/infra/mongobd"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type chatModule struct {
	roomLoaders map[int64]*chatRoomLoader
}

var Chat = systems.RegisterSystem(&chatModule{
	roomLoaders: make(map[int64]*chatRoomLoader),
})

func init() {
	lifecycle.RegisterHandler(actor.CategoryChatMgr, Chat)
}

func (m *chatModule) OnStart() {
}

func (m *chatModule) OnStop() {
}

// OnActorStart 启动行为.
func (m *chatModule) OnActorStart(mgr *actors.ChatMgr) error {
	// 加载或创建服务器聊天室
	if err := m.loadOrCreateServerRoom(mgr); err != nil {
		return err
	}
	return nil
}

// OnActorStop 停机行为.
func (m *chatModule) OnActorStop(mgr *actors.ChatMgr) {
	if err := m.saveDirtyRooms(mgr); err != nil {
		logger.Get().Errorf("[chatModule] save dirty rooms failed, err: %v", err)
	}
}

func (m *chatModule) MgrUID() actor.ActorUID {
	return actor.ActorUID{
		Category: actor.CategoryChatMgr.ActorCategory(),
		ID:       app.Env().ServerID(),
	}
}

// loadOrCreateServerRoom 加载或创建服务器聊天室.
func (m *chatModule) loadOrCreateServerRoom(mgr *actors.ChatMgr) error {
	// 加载服务器聊天室
	room, err := m.loadRoom(mgr, 0)
	if err == nil {
		return nil
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	// 创建服务器聊天室
	room = chat.NewRoom(0, int32(gdconf.ChatRoomTypeServer), time.Now().Unix())
	m.addRoom(mgr, room, true)

	return nil
}

// loadRoom 加载房间.
func (m *chatModule) loadRoom(mgr *actors.ChatMgr, roomId int64) (*chat.Room, error) {
	room := mgr.GetRoom(roomId)
	if room != nil {
		return room, nil
	}

	room = &chat.Room{}

	// 从数据库加载房间.
	op := mongobd.NewOp[mongobd.OpLoad](app.Env().DB(), model.CollChatRoom).
		SetFilter(bson.M{"_id": roomId}).
		SetPrimary(true).
		SetTarget(room)
	if err := app.MongoBD().Exec(time.Now().UnixNano(), op); err != nil {
		return nil, err
	}

	// 添加房间数据
	m.addRoom(mgr, room, false)

	return room, nil
}

// addRoomLoader 添加房间加载器.
func (m *chatModule) addRoomLoader(roomId int64, chatRoomLoader *chatRoomLoader) {
	m.roomLoaders[roomId] = chatRoomLoader
}

// getRoomLoader 获取房间加载器.
func (m *chatModule) getRoomLoader(roomId int64) *chatRoomLoader {
	return m.roomLoaders[roomId]
}

// delRoomLoader 删除房间加载器.
func (m *chatModule) delRoomLoader(roomId int64) {
	delete(m.roomLoaders, roomId)
}

// LoadRoomAsync 异步加载房间.
func (m *chatModule) LoadRoomAsync(ctx any, roomId int64, cb ChatRoomLoadCallback) {
	var mgr *actors.ChatMgr

	switch ctx := ctx.(type) {
	case *actor.Context:
		mgr = actor.CtxActor[*actors.ChatMgr](ctx)
	case *actors.ChatMgr:
		mgr = ctx
	default:
		cb(ctx, nil, fmt.Errorf("loadRoomAsync ctx type %T not supported", ctx))
	}

	room := mgr.GetRoom(roomId)
	if room != nil {
		cb(ctx, room, nil)
		return
	}

	// 获取加载器
	loader := m.getRoomLoader(roomId)

	// 加载器已存在
	if loader != nil && !loader.appendCallback(cb) {
		loader.invokeCallback(ctx, cb)
		return
	}

	// 添加加载器
	loader = newChatRoomLoader(roomId, cb)
	m.addRoomLoader(roomId, loader)
	loader.start(ctx)
}

// addRoom 添加房间.
func (m *chatModule) addRoom(mgr *actors.ChatMgr, room *chat.Room, new bool) {
	roomFixed := room.FixVersion()
	mgr.AddRoom(room)
	if new || roomFixed {
		m.setDirtyRoom(mgr, room)
	}
}

// setDirtyRoom 设置脏数据房间.
func (m *chatModule) setDirtyRoom(mgr *actors.ChatMgr, room *chat.Room) {
	mgr.SetDirtyRoom(room)
	if mgr.DirtyTimerId == gactor.TimerIdNone {
		mgr.DirtyTimerId = mgr.StartTimer(5*time.Second, false, nil, m.onDirtyTimer)
	}
}

// onDirtyTimer 脏数据定时器回调.
func (m *chatModule) onDirtyTimer(args gactor.ActorTimerArgs) {
	mgr := actor.ToActor[*actors.ChatMgr](args.Actor)
	mgr.DirtyTimerId = 0
	if err := m.saveDirtyRooms(mgr); err != nil {
		logger.Get().Errorf("save dirty rooms failed, err: %v", err)
		mgr.DirtyTimerId = mgr.StartTimer(5*time.Second, false, nil, m.onDirtyTimer)
	}
}

// saveDirtyRooms 保存脏数据房间.
func (m *chatModule) saveDirtyRooms(mgr *actors.ChatMgr) error {
	if len(mgr.DirtyRooms) == 0 {
		return nil
	}

	opBulk := mongobd.NewOp[mongobd.OpBulk](app.Env().DB(), model.CollChatRoom)
	opBulk.SetOrdered(true)
	models := make([]mongo.WriteModel, 0, len(mgr.DirtyRooms))
	for _, room := range mgr.DirtyRooms {
		models = append(models, mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": room.ID}).
			SetReplacement(room).
			SetUpsert(true))
	}
	opBulk.SetModels(models)
	if err := app.MongoBD().Exec(time.Now().UnixNano(), opBulk); err != nil {
		return err
	}
	mgr.ClearDirtyRooms()
	return nil
}

// playerAddRoom 玩家添加聊天室.
func (m *chatModule) playerAddRoom(p *actors.Player, room *player.ChatRoom) error {
	chat := actor.GetActorModule[*player.Chat](p, true)
	if chat.GetRoom(room.ID) != nil {
		return actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberAlreadyInRoom)
	}
	chat.AddRoom(room)
	p.SetDirtyModules(chat)
	return nil
}

// playerDelRoom 玩家删除聊天室.
func (m *chatModule) playerDelRoom(p *actors.Player, roomId int64) error {
	chat := actor.GetActorModule[*player.Chat](p, true)
	if chat.GetRoom(roomId) == nil {
		return actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberNotInRoom)
	}
	chat.DelRoom(roomId)
	p.SetDirtyModules(chat)
	return nil
}

type SendMsgParams struct {
	RoomId   int64  // 聊天室ID
	SenderId int64  // 发送者ID
	Content  string // 消息内容
}

// SendMsg 发送聊天消息到参数指定的聊天室.
func (m *chatModule) SendMsg(mgr *actors.ChatMgr, params SendMsgParams, cb func(m *actors.ChatMgr, err error)) {
	m.LoadRoomAsync(mgr, params.RoomId, func(ctx any, room *chat.Room, err error) {
		mgr := ctx.(*actors.ChatMgr)

		if err != nil {
			logger.Get().Errorf("[ChatModule.SendMsg] load room async %d failed, from sender %d, %v", params.RoomId, params.SenderId, err)
			cb(mgr, err)
			return
		}

		// 检查权限
		if err := m.checkSendPriviledge(room, params.SenderId); err != nil {
			cb(mgr, err)
			return
		}

		// 添加消息记录.
		nowTs := time.Now().Unix()
		record := chat.NewMessageRecord(room.NextMsgID(), params.SenderId, params.Content, nowTs)
		room.AddHistory(record, gdconf.Global().GetChatRoomHistoryMaxByType(gdconf.ChatRoomType(room.Type)))
		room.LastActiveAt = nowTs
		m.setDirtyRoom(mgr, room)

		// 转发消息.
		m.forwardMsg(room, record)

		// 成功
		cb(mgr, nil)
	})
}

// checkSendPriviledge 检查发送权限.
func (m *chatModule) checkSendPriviledge(room *chat.Room, senderId int64) error {
	switch gdconf.ChatRoomType(room.Type) {
	case gdconf.ChatRoomTypeServer:
		return nil
	case gdconf.ChatRoomTypeGroup:
		if room.FindMember(senderId) == nil {
			return actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberNotInRoom)
		}
	}
	return nil
}

// forwardMsg 转发消息.
func (m *chatModule) forwardMsg(room *chat.Room, record *chat.MessageRecord) {
	msg := &pbc2s.ChatMsgPush{
		RoomId: room.ID,
		Msg:    m.PackPBHistory(record),
	}

	switch room.Type {
	case int32(gdconf.ChatRoomTypeServer):
		Broadcast.Send2AllOnlinePlayers(msg)
	case int32(gdconf.ChatRoomTypeGroup):
		memberIds := make([]int64, 0, len(room.Members))
		for _, member := range room.Members {
			memberIds = append(memberIds, member.ID)
		}
		Broadcast.Send2OnlinePlayers(msg, memberIds...)
	}
}

// PlayerSendMsg 玩家发送聊天消息.
func (m *chatModule) PlayerSendMsg(p *actors.Player, roomId int64, content string, cb func(p *actors.Player, err error)) {
	if err := p.Sugared().AsyncRPCWithTimeout(m.MgrUID(), &pbs2s.SendChatMsgReq{
		RoomId:   roomId,
		SenderId: p.ID(),
		Content:  content,
	}, func(resp actor.ActorRPCResp) {
		result := actor.NewRPCResult(resp.Reply, resp.Err)
		if !result.Success() {
			logger.Get().Errorf("[chatModule.PlayerSendMsg] player %d, send msg to room %d, %v", p.ID(), roomId, result.Err())
			cb(p, result.Err())
			return
		}
		cb(p, nil)
	}, 5*time.Second); err != nil {
		logger.Get().Errorf("[chatModule.PlayerSendMsg] player %d, asyncRPC send msg to room %d, %v", p.ID(), roomId, err)
		cb(p, actor.WithPbError(pbcommon.ErrCode_ECInternalError))
	}
}

type PlayerCreateRoomParams struct {
	RoomType  int32   // 聊天室类型
	RoomName  string  // 聊天室名称
	MemberIds []int64 // 成员ID列表
}

// PlayerCreateRoom 玩家创建聊天室.
func (m *chatModule) PlayerCreateRoom(p *actors.Player, params PlayerCreateRoomParams, cb func(p *actors.Player, err error)) {
	// 参数检查
	switch gdconf.ChatRoomType(params.RoomType) {
	case gdconf.ChatRoomTypeGroup:
		// 检查成员
		memberMap := make(map[int64]struct{}, len(params.MemberIds))
		for _, memberId := range params.MemberIds {
			if memberId == p.ID() {
				cb(p, actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberIsSelf))
				return
			}
			if _, ok := memberMap[memberId]; ok {
				cb(p, actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberDuplicate))
				return
			}
			memberMap[memberId] = struct{}{}
		}
	default:
		cb(p, actor.WithPbError(pbcommon.ErrCode_ECChatRoomTypeInvalid))
		return
	}

	// 创建房间
	switch gdconf.ChatRoomType(params.RoomType) {
	case gdconf.ChatRoomTypeGroup:
		m.playerDoCreateGroupChatRoom(p, params, cb)
		return
	}
}

// playerDoCreateGroupChatRoom 玩家创建群聊室.
func (m *chatModule) playerDoCreateGroupChatRoom(p *actors.Player, params PlayerCreateRoomParams, cb func(p *actors.Player, err error)) {
	// 异步创建聊天室
	if err := p.Sugared().AsyncRPCWithTimeout(m.MgrUID(), &pbs2s.CreateGroupChatRoomReq{
		Name:      params.RoomName,
		MemberIds: append([]int64{p.ID()}, params.MemberIds...),
	}, func(resp actor.ActorRPCResp) {
		result := actor.NewRPCResult(resp.Reply, resp.Err)
		if !result.Success() {
			cb(p, result.Err())
			return
		}
		cb(p, nil)
	}, 5*time.Second); err != nil {
		logger.Get().Errorf("[chatModule.PlayerDoCreateGroupChatRoom] player %d, asyncRPC create group chat room, %v", p.ID(), err)
		cb(p, actor.WithPbError(pbcommon.ErrCode_ECInternalError))
	}
}

type CreateGroupChatRoomParams struct {
	Name      string  // 群聊室名称
	MemberIds []int64 // 成员ID列表
}

// CreateGroupChatRoom 创建群聊室.
func (m *chatModule) CreateGroupChatRoom(mgr *actors.ChatMgr, params *CreateGroupChatRoomParams, cb func(mgr *actors.ChatMgr, err error)) {
	// 检查成员
	if len(params.MemberIds) == 0 {
		cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberEmpty))
		return
	}
	if len(params.MemberIds) > int(gdconf.Global().GetChatRoomMemberMaxByType(gdconf.ChatRoomTypeGroup)) {
		cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberCountExceed))
		return
	}
	memberMap := make(map[int64]struct{}, len(params.MemberIds))
	for _, memberId := range params.MemberIds {
		if _, ok := memberMap[memberId]; ok {
			cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberDuplicate))
			return
		}
		memberMap[memberId] = struct{}{}
	}

	// 创建聊天室
	ownerId := params.MemberIds[0]
	room := chat.NewRoom(mgr.NextRoomID(), int32(gdconf.ChatRoomTypeGroup), time.Now().Unix())
	room.Name = params.Name
	room.OwnerID = ownerId
	room.AddMember(chat.NewRoomMember(ownerId))

	// 通知群主加入聊天室
	if err := mgr.Sugared().AsyncRPCWithTimeout(actor.PlayerUID(ownerId), &pbs2s.NotifyJoinChatRoomReq{
		RoomId:   room.ID,
		RoomType: room.Type,
	}, func(resp actor.ActorRPCResp) {
		chatMgr := actor.ToActor[*actors.ChatMgr](resp.Actor)

		result := actor.NewRPCResult(resp.Reply, resp.Err)
		if !result.Success() {
			logger.Get().Errorf("[ChatModule] notify owner %d to join group room, %v", ownerId, result.Err())
			cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomCreateFailed))
			return
		}

		// 添加聊天室
		m.addRoom(chatMgr, room, true)

		// 回复
		cb(mgr, nil)
		Player.Send2OnlinePlayer(ownerId, &pbc2s.JoinChatRoomPush{
			Room: m.packPBRoom(room),
		})

		// 异步通知其它成员加入聊天室.
		for _, memberId := range params.MemberIds[1:] {
			m.notifyMemberJoinChatRoom(chatMgr, room, memberId)
		}

	}, 5*time.Second); err != nil {
		logger.Get().Errorf("[ChatModule] notify owner %d to join group room, %v", ownerId, err)
		cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomCreateFailed))
	}
}

// notifyMemberJoinChatRoom 通知成员加入聊天室.
func (m *chatModule) notifyMemberJoinChatRoom(chatMgr *actors.ChatMgr, room *chat.Room, memberId int64) {
	roomId := room.ID
	if err := chatMgr.Sugared().AsyncRPCWithTimeout(actor.PlayerUID(memberId), &pbs2s.NotifyJoinChatRoomReq{
		RoomId:   room.ID,
		RoomType: room.Type,
	},
		func(resp actor.ActorRPCResp) {
			result := actor.NewRPCResult(resp.Reply, resp.Err)
			if !result.Success() {
				logger.Get().Errorf("[ChatModule.notifyMemberJoinChatRoom] notify member %d to join group room, %v", memberId, result.Err())
				return
			}
			chatMgr := actor.ToActor[*actors.ChatMgr](resp.Actor)
			m.LoadRoomAsync(chatMgr, roomId, func(ctx any, room *chat.Room, err error) {
				chatMgr := ctx.(*actors.ChatMgr)
				if err != nil {
					logger.Get().Errorf("[ChatModule.notifyMemberJoinChatRoom] load room async %d failed, member %d, %v", roomId, memberId, err)
					return
				}
				maxMemebers := gdconf.Global().GetChatRoomMemberMaxByType(gdconf.ChatRoomType(room.Type))
				if room.MemberCount() >= maxMemebers {
					// 聊天室已满.
					if err := chatMgr.Sugared().AsyncRPCWithTimeout(actor.PlayerUID(memberId), &pbs2s.NotifyLeaveChatRoomReq{
						RoomId: roomId,
					}, func(resp actor.ActorRPCResp) {
						result := actor.NewRPCResult(resp.Reply, resp.Err)
						if !result.Success() {
							logger.Get().Errorf("[ChatModule.notifyMemberJoinChatRoom] room full, notify member %d to leave group room, %v", memberId, result.Err())
							return
						}
					}, 5*time.Second); err != nil {
						logger.Get().Errorf("[ChatModule.notifyMemberJoinChatRoom] room full, async notify member %d to leave group room, %v", memberId, err)
						return
					}
					return
				}
				room.AddMember(chat.NewRoomMember(memberId))
				m.setDirtyRoom(chatMgr, room)
				Player.Send2OnlinePlayer(memberId, &pbc2s.JoinChatRoomPush{
					Room: m.packPBRoom(room),
				})
				m.broadcastMemberJoinRoom(room, memberId)
			})
		}, 5*time.Second); err != nil {
		logger.Get().Errorf("[ChatModule] notify member %d to join group room, %v", memberId, err)
	}
}

// broadcastMemberJoinRoom 广播成员加入聊天室.
func (m *chatModule) broadcastMemberJoinRoom(room *chat.Room, memberId int64) {
	member := room.FindMember(memberId)
	if member == nil {
		return
	}
	if room.MemberCount() == 1 {
		return
	}
	memberIds := make([]int64, 0, room.MemberCount()-1)
	for _, m := range room.Members {
		if m.ID != memberId {
			memberIds = append(memberIds, m.ID)
		}
	}
	Broadcast.Send2OnlinePlayers(&pbc2s.ChatRoomMemberJoinPush{
		RoomId:     room.ID,
		NewMembers: []*pbcommon.ChatRoomMember{&pbcommon.ChatRoomMember{Id: memberId}},
	}, memberIds...)
}

// PlayerAddRoom 玩家添加聊天室.
func (m *chatModule) PlayerAddRoom(p *actors.Player, roomId int64, roomType int32) error {
	room := player.NewChatRoom(roomId, roomType)
	return m.playerAddRoom(p, room)
}

// PlayerLeaveChatRoom 玩家退出聊天室.
func (m *chatModule) PlayerLeaveChatRoom(p *actors.Player, roomId int64, cb func(p *actors.Player, err error)) {
	chat := actor.GetActorModule[*player.Chat](p, true)
	room := chat.GetRoom(roomId)
	if room == nil {
		cb(p, actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberNotInRoom))
		return
	}
	if room.Type != int32(gdconf.ChatRoomTypeGroup) {
		cb(p, actor.WithPbError(pbcommon.ErrCode_ECChatRoomCantLeave))
		return
	}
	if err := p.Sugared().AsyncRPCWithTimeout(m.MgrUID(), &pbs2s.LeaveChatRoomReq{RoomId: roomId, MemberId: p.ID()},
		func(resp actor.ActorRPCResp) {
			result := actor.NewRPCResult(resp.Reply, resp.Err)
			if !result.Success() {
				logger.Get().Errorf("[ChatModule.PlayerLeaveChatRoom] player %d leave room %d failed, %v", p.ID(), roomId, result.Err())
				cb(p, result.Err())
				return
			}
			cb(p, nil)
		}, 5*time.Second); err != nil {
		logger.Get().Errorf("[ChatModule.PlayerLeaveChatRoom] player %d, asyncRPC leave room %d failed, %v", p.ID(), roomId, err)
		cb(p, actor.WithPbError(pbcommon.ErrCode_ECInternalError))
	}
}

// checkLeaveRoom 检查退出聊天室.
func (m *chatModule) checkLeaveRoom(room *chat.Room, memberId int64) error {
	switch room.Type {
	case int32(gdconf.ChatRoomTypeGroup):
		if memberId == room.OwnerID {
			return actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberIsOwner)
		}
		if room.FindMember(memberId) == nil {
			return actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberNotInRoom)
		}
		return nil
	default:
		return actor.WithPbError(pbcommon.ErrCode_ECChatRoomCantLeave)
	}
}

// LeaveChatRoom 退出聊天室.
func (m *chatModule) LeaveChatRoom(mgr *actors.ChatMgr, roomId int64, memberId int64, cb func(mgr *actors.ChatMgr, err error)) {
	m.LoadRoomAsync(mgr, roomId, func(ctx any, room *chat.Room, err error) {
		if err != nil {
			logger.Get().Errorf("[ChatModule.LeaveChatRoom] member %d, load room %d failed, %v", memberId, roomId, err)
			cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomNotExists))
			return
		}

		// 检查退出聊天室.
		if err := m.checkLeaveRoom(room, memberId); err != nil {
			cb(mgr, err)
			return
		}

		// 通知成员退出聊天室.
		if err := mgr.Sugared().AsyncRPCWithTimeout(actor.PlayerUID(memberId), &pbs2s.NotifyLeaveChatRoomReq{RoomId: room.ID},
			func(resp actor.ActorRPCResp) {
				mgr := actor.ToActor[*actors.ChatMgr](resp.Actor)

				result := actor.NewRPCResult(resp.Reply, resp.Err)
				if !result.Success() {
					err := result.Err()
					var pbErr *actor.PbError
					if !(errors.As(err, &pbErr) && pbErr.Err.Code == pbcommon.ErrCode_ECChatRoomMemberNotInRoom) {
						logger.Get().Errorf("[ChatModule.LeaveChatRoom] notify member %d to leave room %d failed, %v", memberId, roomId, result.Err())
						cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberNotInRoom))
						return
					}
				}

				// 回复成功
				cb(mgr, nil)

				// 加载聊天室数据以删除成员
				Chat.LoadRoomAsync(mgr, roomId, func(ctx any, room *chat.Room, err error) {
					mgr := ctx.(*actors.ChatMgr)
					if err != nil {
						logger.Get().Errorf("[ChatModule.LeaveChatRoom] after notify member %d to leave, load room %d failed, %v", memberId, roomId, err)
						return
					}
					// 从聊天室中删除成员
					room.RemoveMember(memberId)
					Chat.setDirtyRoom(mgr, room)
					Chat.broadcastMemberLeaveRoom(room, memberId)
				})
			}, 5*time.Second); err != nil {
			logger.Get().Errorf("[ChatModule.LeaveChatRoom] notify member %d to leave room %d failed, %v", memberId, roomId, err)
			cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECInternalError))
			return
		}
	})
}

// broadcastMemberLeaveRoom 广播成员退出聊天室.
func (m *chatModule) broadcastMemberLeaveRoom(room *chat.Room, memberId int64) {
	memberIds := make([]int64, 0, room.MemberCount())
	for _, m := range room.Members {
		memberIds = append(memberIds, m.ID)
	}
	Broadcast.Send2OnlinePlayers(&pbc2s.ChatRoomMemberLeavePush{
		RoomId:    room.ID,
		MemberIds: []int64{memberId},
	}, memberIds...)
}

// PlayerDelRoom 玩家删除聊天室.
func (m *chatModule) PlayerDelRoom(p *actors.Player, roomId int64) error {
	return m.playerDelRoom(p, roomId)
}

// PlayerChatRoomInvite 玩家邀请玩家加入聊天室.
func (m *chatModule) PlayerChatRoomInvite(p *actors.Player, roomId int64, targetIds []int64, cb func(p *actors.Player, err error)) {
	chat := actor.GetActorModule[*player.Chat](p, true)
	if chat.GetRoom(roomId) == nil {
		cb(p, actor.WithPbError(pbcommon.ErrCode_ECChatRoomNotExists))
		return
	}
	if len(targetIds) == 0 {
		cb(p, actor.WithPbError(pbcommon.ErrCode_ECChatRoomInviteTargetEmpty))
		return
	}
	targetMap := make(map[int64]struct{}, len(targetIds))
	for _, id := range targetIds {
		if _, ok := targetMap[id]; ok {
			cb(p, actor.WithPbError(pbcommon.ErrCode_ECChatRoomInviteTargetDuplicate))
			return
		}
		targetMap[id] = struct{}{}
	}
	if err := p.Sugared().AsyncRPCWithTimeout(m.MgrUID(), &pbs2s.ChatRoomInviteReq{
		RoomId:    roomId,
		InviterId: p.ID(),
		TargetIds: targetIds,
	}, func(resp actor.ActorRPCResp) {
		result := actor.NewRPCResult(resp.Reply, resp.Err)
		if !result.Success() {
			err := result.Err()
			logger.Get().Errorf("[ChatModule.PlayerChatRoomInvite] player %d invite %+v to room %d failed, %v", p.ID(), targetIds, roomId, result.Err())
			cb(p, err)
			return
		}
		cb(p, nil)
	}, 5*time.Second); err != nil {
		logger.Get().Errorf("[ChatModule.PlayerChatRoomInvite] player %d invite %+v to room %d, async call mgr failed, %v", p.ID(), targetIds, roomId, err)
		cb(p, actor.WithPbError(pbcommon.ErrCode_ECInternalError))
	}
}

type ChatRoomInviteParams struct {
	RoomId    int64   // 聊天室ID
	InviterId int64   // 邀请人ID
	TargetIds []int64 // 邀请目标ID列表
}

// ChatRoomInvite 邀请玩家加入聊天室.
func (m *chatModule) ChatRoomInvite(mgr *actors.ChatMgr, params *ChatRoomInviteParams, cb func(mgr *actors.ChatMgr, err error)) {
	m.LoadRoomAsync(mgr, params.RoomId, func(ctx any, room *chat.Room, err error) {
		mgr := ctx.(*actors.ChatMgr)
		if err != nil {
			logger.Get().Errorf("[ChatModule.ChatRoomInvite] %d invite %+v, load room %d failed, %v", params.InviterId, params.TargetIds, params.RoomId, err)
			cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomNotExists))
			return
		}
		if room.Type != int32(gdconf.ChatRoomTypeGroup) {
			cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomTypeCantInvite))
			return
		}
		if params.InviterId != room.OwnerID {
			cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomNoPriviledge))
			return
		}
		maxMembers := gdconf.Global().GetChatRoomMemberMaxByType(gdconf.ChatRoomType(room.Type))
		if room.MemberCount()+int32(len(params.TargetIds)) > maxMembers {
			cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberCountExceed))
			return
		}
		for _, targetId := range params.TargetIds {
			if room.FindMember(targetId) != nil {
				cb(mgr, actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberAlreadyInRoom))
				return
			}
		}
		for _, targetId := range params.TargetIds {
			m.notifyMemberJoinChatRoom(mgr, room, targetId)
		}
		cb(mgr, nil)
	})
}

// CheckChatHistory 检查聊天历史权限.
func (m *chatModule) CheckChatHistoryPriviledge(room *chat.Room, memberId int64) error {
	switch gdconf.ChatRoomType(room.Type) {
	case gdconf.ChatRoomTypeServer:
		return nil
	case gdconf.ChatRoomTypeGroup:
		if room.FindMember(memberId) == nil {
			return actor.WithPbError(pbcommon.ErrCode_ECChatRoomMemberNotInRoom)
		}
	}
	return nil
}

// GetRoomHistory 获取聊天历史.
func (m *chatModule) GetRoomHistory(room *chat.Room, lastMsgId int64) []*chat.MessageRecord {
	head := 0
	tail := len(room.Histories)
	if lastMsgId > 0 {
		tail = room.IndexHistory(lastMsgId)
	}
	head = int(math.Max(0, float64(tail)-10))
	return room.Histories[head:tail]
}

// packPBRoom 打包聊天室.
func (m *chatModule) packPBRoom(room *chat.Room) *pbcommon.ChatRoom {
	pb := &pbcommon.ChatRoom{
		Base: &pbcommon.ChatRoomBase{
			Id:      room.ID,
			Type:    room.Type,
			Name:    room.Name,
			OwnerId: room.OwnerID,
		},
	}
	pb.Members = make([]*pbcommon.ChatRoomMember, len(room.Members))
	for i, m := range room.Members {
		pb.Members[i] = &pbcommon.ChatRoomMember{
			Id: m.ID,
		}
	}
	return pb
}

// PackPBHistory 打包聊天历史.
func (m *chatModule) PackPBHistory(history *chat.MessageRecord) *pbcommon.ChatMsg {
	return &pbcommon.ChatMsg{
		Id:        history.ID,
		SenderId:  history.SenderID,
		Content:   history.Content,
		CreatedAt: history.CreatedAt,
	}
}

// PackPBHistorys 打包聊天历史记录.
func (m *chatModule) PackPBHistorys(records []*chat.MessageRecord) []*pbcommon.ChatMsg {
	msgs := make([]*pbcommon.ChatMsg, 0, len(records))
	for _, r := range records {
		msgs = append(msgs, m.PackPBHistory(r))
	}
	return msgs
}

// ChatRoomLoadCallback 聊天室加载回调.
type ChatRoomLoadCallback func(ctx any, room *chat.Room, err error)

const (
	chatRoomLoaderStateNone = 0
	chatRoomLoaderStateInit = 1
	chatRoomLoaderStateWait = 2
	chatRoomLoaderStateEnd  = 3
)

// chatRoomLoader 聊天室加载器.
type chatRoomLoader struct {
	mtx         sync.Mutex             // 互斥锁
	state       int32                  // 状态
	initCond    sync.Cond              // 初始化完成条件
	roomId      int64                  // 聊天室ID
	room        *chat.Room             // 聊天室数据
	callbacks   []ChatRoomLoadCallback // 结果回调
	asyncCaller actor.ActorAsyncCaller // 异步调用器
	err         error                  // 错误
}

func newChatRoomLoader(roomId int64, callbacks ...ChatRoomLoadCallback) *chatRoomLoader {
	l := &chatRoomLoader{
		roomId:    roomId,
		callbacks: callbacks,
	}
	l.initCond.L = &l.mtx
	return l
}

// start 开始加载.
func (l *chatRoomLoader) start(ctx any) {
	l.mtx.Lock()
	if l.state != chatRoomLoaderStateNone {
		l.mtx.Unlock()
		return
	}
	l.state = chatRoomLoaderStateInit
	l.mtx.Unlock()

	// 异步加载房间数据.
	room := &chat.Room{}
	op := mongobd.NewOp[mongobd.OpLoad](app.Env().DB(), model.CollChatRoom).
		SetFilter(bson.M{"_id": l.roomId}).
		SetPrimary(true).
		SetTarget(room)
	if err := app.MongoBD().Add(l.roomId, op, l.loadCallback); err != nil {
		l.endWithErr(ctx, err)
		return
	}

	// 添加异步调用.
	asyncCaller, err := l.asyncCall(ctx)
	if err != nil {
		l.endWithErr(ctx, err)
		return
	}

	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.room = room
	l.asyncCaller = asyncCaller
	l.state = chatRoomLoaderStateWait
	l.initCond.Signal()
}

func (l *chatRoomLoader) endWithErr(ctx any, err error) {
	l.mtx.Lock()
	if l.state == chatRoomLoaderStateEnd {
		l.mtx.Unlock()
		return
	}
	l.err = err
	l.room = nil
	l.state = chatRoomLoaderStateEnd
	l.mtx.Unlock()
	l.invokeCallbacks(ctx)
}

// loadCallback 加载回调.
func (l *chatRoomLoader) loadCallback(op mongobd.Op) {
	l.mtx.Lock()
	if l.state == chatRoomLoaderStateInit {
		l.initCond.Wait()
	}
	if err := op.Err(); err != nil {
		l.err = err
		l.room = nil
	}
	l.mtx.Unlock()
	l.asyncCaller(nil, nil)
}

func (l *chatRoomLoader) asyncCall(ctx any) (actor.ActorAsyncCaller, error) {
	switch ctx := ctx.(type) {
	case *actor.Context:
		return ctx.AsyncCall(l.asyncCallback1, 5*time.Second)
	case *actors.ChatMgr:
		return ctx.AsyncCall(l.asyncCallback2, 5*time.Second)
	default:
		return nil, errors.New("invalid context")
	}
}

// asyncCallback1 异步回调.
func (l *chatRoomLoader) asyncCallback1(ctx *actor.Context, args any, err error) {
	l.mtx.Lock()
	if l.state == chatRoomLoaderStateEnd {
		l.mtx.Unlock()
		return
	}
	mgr := actor.CtxActor[*actors.ChatMgr](ctx)
	if l.room != nil {
		Chat.addRoom(mgr, l.room, false)
	}
	l.state = chatRoomLoaderStateEnd
	l.mtx.Unlock()
	l.invokeCallbacks(ctx)
	Chat.delRoomLoader(l.roomId)
}

// asyncCallback2 异步回调.
func (l *chatRoomLoader) asyncCallback2(ctx actor.Actor, args any, err error) {
	l.mtx.Lock()
	if l.state == chatRoomLoaderStateEnd {
		l.mtx.Unlock()
		return
	}
	mgr := actor.ToActor[*actors.ChatMgr](ctx)
	if l.room != nil {
		Chat.addRoom(mgr, l.room, false)
	}
	l.state = chatRoomLoaderStateEnd
	l.mtx.Unlock()
	l.invokeCallbacks(ctx)
	Chat.delRoomLoader(l.roomId)
}

// appendCallback 添加回调.
// 若加载起已完成, 失败.
func (l *chatRoomLoader) appendCallback(cb ChatRoomLoadCallback) bool {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if l.state == chatRoomLoaderStateEnd {
		return false
	}
	l.callbacks = append(l.callbacks, cb)
	return true
}

// invokeCallbacks 执行回调函数.
func (l *chatRoomLoader) invokeCallbacks(ctx any) {
	for _, cb := range l.callbacks {
		cb(ctx, l.room, l.err)
	}
}

// invokeCallback 执行指定回调函数.
func (l *chatRoomLoader) invokeCallback(ctx any, cb ChatRoomLoadCallback) {
	cb(ctx, l.room, l.err)
}
