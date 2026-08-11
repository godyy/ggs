package gdconf

// GetChatRoomHistoryMaxByType 获取聊天房间历史消息最大数量.
// 若未配置，返回默认值.
func (t *global) GetChatRoomHistoryMaxByType(roomType ChatRoomType) int32 {
	max, ok := t.ChatRoomTypeHistoryMax[roomType]
	if ok {
		return max
	}
	return t.DefaultChatRoomHistoryMax
}

// GetChatRoomMemberMaxByType 获取聊天房间成员最大数量.
func (t *global) GetChatRoomMemberMaxByType(roomType ChatRoomType) int32 {
	if t.ChatRoomMemberMax == nil {
		return 0
	}
	return t.ChatRoomMemberMax[roomType]
}
