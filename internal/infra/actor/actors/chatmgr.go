package actors

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/gdconf"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/lifecycle"
	"github.com/godyy/ggs/internal/infra/actor/model/chat"
	"github.com/godyy/gutils/container/lru"
)

// ChatMgr 聊天管理Actor.
type ChatMgr struct {
	actor.ActorWithModel[*chat.Mgr] // 集成携带数据模型的Actor封装

	ServerRoom   *chat.Room                  // 服务器房间.
	CachedRooms  *lru.LRU[int64, *chat.Room] // 已缓存的房间数据.
	DirtyRooms   map[int64]*chat.Room        // 脏数据的房间.
	DirtyTimerId actor.TimerId               // 脏数据定时器ID.
}

// NewChatMgr 构造聊天管理Actor.
func NewChatMgr(a actor.Actor) *ChatMgr {
	return &ChatMgr{
		ActorWithModel: actor.NewActorWithModel[*chat.Mgr](a),
		CachedRooms:    lru.New[int64, *chat.Room](1000),
		DirtyRooms:     map[int64]*chat.Room{},
	}
}

// OnStart 启动行为.
func (cm *ChatMgr) OnStart() error {
	cm.Model = chat.NewMgr()
	if err := cm.ActorWithModel.OnStart(); err != nil {
		return err
	}
	return lifecycle.OnStart(cm)
}

// OnStop 停机行为.
func (cm *ChatMgr) OnStop() error {
	lifecycle.OnStop(cm)
	return cm.ActorWithModel.OnStop()
}

// NextRoomID 获取下一个房间ID.
func (cm *ChatMgr) NextRoomID() int64 {
	roomId := cm.Model.NextRoomID()
	cm.SetAllDirty()
	return roomId
}

// AddRoom 添加房间数据.
func (cm *ChatMgr) AddRoom(room *chat.Room) {
	if room.Type == int32(gdconf.ChatRoomTypeServer) {
		cm.ServerRoom = room
	} else {
		cm.CachedRooms.Put(room)
	}
}

// GetRoom 获取房间数据.
// 若缓存未命中，但是命中了脏数据，会将脏数据重新刷新到缓存中.
func (cm *ChatMgr) GetRoom(roomId int64) *chat.Room {
	// 匹配服务器聊天室
	if cm.ServerRoom != nil && roomId == cm.ServerRoom.ID {
		return cm.ServerRoom
	}

	// 匹配已缓存的房间
	room, exist := cm.CachedRooms.Get(roomId)
	if exist {
		return room
	}

	// 匹配脏数据的房间
	room, exist = cm.DirtyRooms[roomId]
	if exist {
		cm.CachedRooms.Put(room)
		return room
	}

	return nil
}

// SetDirtyRoom 设置脏数据的房间.
func (cm *ChatMgr) SetDirtyRoom(room *chat.Room) {
	cm.DirtyRooms[room.ID] = room
}

// ClearDirtyRooms 清除脏数据的房间.
func (cm *ChatMgr) ClearDirtyRooms() {
	cm.DirtyRooms = make(map[int64]*chat.Room)
}

func init() {
	registerDefine(gactor.NewActorDefine(gactor.ActorDefineConfig{
		Name:              actor.CategoryChatMgr.String(),
		Category:          actor.CategoryChatMgr.ActorCategory(),
		Priority:          0,
		PriMessageBoxSize: 5000,
		MessageBoxSize:    5000,
		BehaviorCreator: func(a gactor.Actor) gactor.ActorBehavior {
			return NewChatMgr(a)
		},
	},
	))
}
