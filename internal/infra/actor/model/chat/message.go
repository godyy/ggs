package chat

// MessageRecord 消息记录.
type MessageRecord struct {
	ID        int64  `bson:"id"`         // 消息的ID.
	SenderID  int64  `bson:"sender_id"`  // 发送者ID.
	Content   string `bson:"content"`    // 消息内容.
	CreatedAt int64  `bson:"created_at"` // 创建时间.
}

func NewMessageRecord(id int64, senderID int64, content string, createdAt int64) *MessageRecord {
	return &MessageRecord{
		ID:        id,
		SenderID:  senderID,
		Content:   content,
		CreatedAt: createdAt,
	}
}
