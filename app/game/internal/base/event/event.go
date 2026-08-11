package event

import "github.com/godyy/gevent"

// ID 事件ID
type ID = gevent.EventID[Kind, int64]

// LID 事件监听ID
type LID = gevent.LID

// MakeID 创建事件ID
func MakeID(kind Kind, Value int64) ID {
	return ID{Kind: kind, Value: Value}
}

// MakeKindID 创建时间类型ID
func MakeKindID(kind Kind) ID {
	return ID{Kind: kind, Value: 0}
}
