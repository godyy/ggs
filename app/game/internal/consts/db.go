package consts

import "time"

const (
	MgoDBCollPlayers     = "players"
	MgoDBCollGameServers = "gameservers"
)

// DirtyPersistDelay 脏数据持久化延迟
const DirtyPersistDelay = 5 * time.Second
