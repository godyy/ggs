package actor

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggskit/infra/actor"
)

// ActorUID Actor唯一标识类型映射.
type ActorUID = actor.ActorUID

// Actor Actor类型映射.
type Actor = actor.Actor

// CActor CActor类型映射.
type CActor = actor.CActor

// TimerId TimerId类型映射.
type TimerId = gactor.TimerId

// Client Actor客户端类型映射.
type Client = actor.Client

// Service Actor服务类型映射.
type Service = actor.Service

// Codec Actor编码器类型映射.
type Codec = actor.Codec

// Registry Actor注册表类型映射.
type Registry = actor.Registry

// ServerStore Actor服务器存储类型映射.
type ServerStore = actor.ServerStore

// Router Actor路由类型映射.
type Router = actor.Router

// Context Actor上下文类型映射.
type Context = actor.Context
