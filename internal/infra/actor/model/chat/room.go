package chat

import (
	"sort"
)

const (
	RoomVersionInit         = 0 // 初始版本号.
	RoomVersionAddHistoryID = 1 // 历史记录ID号.
	RoomVersionCur          = RoomVersionAddHistoryID
)

// Room Room 数据模型.
type Room struct {
	ID           int64            `bson:"_id"`            // 房间ID.
	Type         int32            `bson:"type"`           // 房间类型.
	OwnerID      int64            `bson:"owner_id"`       // 房间所有者ID.
	CreatedAt    int64            `bson:"created_at"`     // 房间创建时间.
	LastActiveAt int64            `bson:"last_active_at"` // 房间最后活跃时间.
	Name         string           `bson:"name"`           // 房间名称.
	Members      []*RoomMember    `bson:"members"`        // 房间成员.
	MsgIDIncr    int64            `bson:"msg_id_incr"`    // 消息ID递增.
	Histories    []*MessageRecord `bson:"histories"`      // 房间消息记录.
	Version      int32            `bson:"version"`        // 版本号.
}

func NewRoom(id int64, typ int32, nowTs int64) *Room {
	return &Room{
		ID:           id,
		Type:         typ,
		CreatedAt:    nowTs,
		LastActiveAt: nowTs,
		Version:      RoomVersionCur,
	}
}

// LRUKey LRU 键.
func (r *Room) LRUKey() int64 {
	return r.ID
}

// AddMember 添加房间成员.
func (r *Room) AddMember(member *RoomMember) bool {
	if r.FindMember(member.ID) != nil {
		return false
	}
	r.Members = append(r.Members, member)
	return true
}

// FindMember 查找房间成员.
func (r *Room) FindMember(memberID int64) *RoomMember {
	for _, member := range r.Members {
		if member.ID == memberID {
			return member
		}
	}
	return nil
}

// RemoveMember 移除房间成员.
func (r *Room) RemoveMember(memberID int64) bool {
	for i, member := range r.Members {
		if member.ID == memberID {
			r.Members = append(r.Members[:i], r.Members[i+1:]...)
			return true
		}
	}
	return false
}

// MemberCount 房间成员数量.
func (r *Room) MemberCount() int32 {
	return int32(len(r.Members))
}

// NextMsgID 获取下一个消息ID.
func (r *Room) NextMsgID() int64 {
	r.MsgIDIncr++
	return r.MsgIDIncr
}

// AddHistory 添加历史记录.
func (r *Room) AddHistory(record *MessageRecord, max int32) {
	l := int32(len(r.Histories))
	if l >= max {
		rem := l - max + 1
		copy(r.Histories, r.Histories[rem:])
		r.Histories = r.Histories[:l-rem]
	}
	r.Histories = append(r.Histories, record)
}

// IndexHistory 查找历史记录索引.
func (r *Room) IndexHistory(msgID int64) int {
	return sort.Search(len(r.Histories), func(i int) bool {
		return r.Histories[i].ID >= msgID
	})
}

// FixData 修复房间数据.
func (r *Room) FixVersion() bool {
	fixed := false
	for r.Version < RoomVersionCur {
		r.Version++
		f := fixVersionFuncs[r.Version]
		if f == nil {
			continue
		}
		if f(r) {
			fixed = true
		}
	}
	return fixed
}

// fixVersionFuncs 版本号修复函数.
var fixVersionFuncs = map[int32]func(room *Room) bool{
	RoomVersionAddHistoryID: fixRoomAddHistoryID,
}

func fixRoomAddHistoryID(room *Room) bool {
	if room.MsgIDIncr == 0 && len(room.Histories) == 0 {
		return false
	}

	for _, record := range room.Histories {
		record.ID = room.NextMsgID()
	}
	return true
}
