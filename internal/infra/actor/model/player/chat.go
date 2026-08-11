package player

// ChatRoom 聊天室.
type ChatRoom struct {
	ID   int64 `bson:"id"`   // 聊天室ID.
	Type int32 `bson:"type"` // 聊天室类型.
}

func NewChatRoom(id int64, t int32) *ChatRoom {
	return &ChatRoom{
		ID:   id,
		Type: t,
	}
}

// Chat 玩家聊天数据.
type Chat struct {
	Rooms map[int64]*ChatRoom `bson:"rooms"` // 已加入的聊天室
}

func (c *Chat) OnInit() {
	c.Rooms = make(map[int64]*ChatRoom)
}

func (c *Chat) ModuleKey() string { return "chat" }

// GetRoom 获取聊天室.
func (c *Chat) GetRoom(roomId int64) *ChatRoom {
	return c.Rooms[roomId]
}

// AddRoom 添加聊天室.
func (c *Chat) AddRoom(room *ChatRoom) bool {
	if c.GetRoom(room.ID) != nil {
		return false
	}
	c.Rooms[room.ID] = room
	return true
}

// DelRoom 删除聊天室.
func (c *Chat) DelRoom(roomId int64) bool {
	if c.GetRoom(roomId) == nil {
		return false
	}
	delete(c.Rooms, roomId)
	return true
}
