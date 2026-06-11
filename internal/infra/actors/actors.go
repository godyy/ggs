package actors

import (
	"time"

	"github.com/godyy/ggskit/infra/actor"
)

const (
	ActorSaveDelay = 5 * time.Second // Actor 存储延迟.
)

// GetModule 通过actor获取模块的通用泛型封装.
func GetModule[M actor.Module](a actor.ActorWithModule, autoCreate bool) M {
	return actor.GetModuleOfActor[M](a, autoCreate)
}
