package chat

// RoomMember 聊天室成员.
type RoomMember struct {
	ID int64 `bson:"id"` // 成员ID.
}

func NewRoomMember(id int64) *RoomMember {
	return &RoomMember{
		ID: id,
	}
}
