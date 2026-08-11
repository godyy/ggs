package gevent

import "math"

// EventID 事件ID，用于唯一标识一个事
// EK 事件类型
// EV 事件值类型
type EventID[EK, EV comparable] struct {
	Kind  EK
	Value EV
}

// LID 监听ID
// 用于唯一标识一个监听
type LID = int32

// maxLID 最大监听ID
// 监听ID必须在 [minLID, maxLID] 范围内
const maxLID = math.MaxInt32

// Event 产生的事件
type Event[EK, EV comparable] struct {
	eventID EventID[EK, EV] // 事件ID
	lid     LID             // 监听ID
	creator any             // 构造者
	params  any             // 参数
}

// EventID 事件ID
func (e *Event[EK, EV]) EventID() EventID[EK, EV] { return e.eventID }

// LID 监听ID
func (e *Event[EK, EV]) LID() LID { return e.lid }

// Creator 构造者
func (e *Event[EK, EV]) Creator() any { return e.creator }

// Params 参数
func (e *Event[EK, EV]) Params() any { return e.params }
