package event

// Kind 事件类型
type Kind = int32

const (
	_                 = Kind(iota)
	KindPlayerOnline  // 玩家上线
	KindPlayerOffline // 玩家离线
)
