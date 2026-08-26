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

// ActorBehavior Actor行为类型映射.
type ActorBehavior = actor.ActorBehavior

// CActorBehavior CActor行为类型映射.
type CActorBehavior = actor.CActorBehavior

// TimerId TimerId类型映射.
type TimerId = gactor.TimerId

// Client Actor客户端类型映射.
type Client = actor.Client

// Service Actor服务类型映射.
type Service = actor.Service

// Codec Actor编码器类型映射.
type Codec = actor.Codec

// RegistryConfig Actor注册表配置类型映射.
type RegistryConfig = actor.RegistryConfig

// Registry Actor注册表类型映射.
type Registry = actor.Registry

// ServerStoreConfig Actor服务器存储配置类型映射.
type ServerStoreConfig = actor.ServerStoreConfig

// ServerStore Actor服务器存储类型映射.
type ServerStore = actor.ServerStore

// Router Actor路由类型映射.
type Router = actor.Router

// Context Actor上下文类型映射.
type Context = actor.Context

// HandlerFunc Actor处理函数类型映射.
type HandlerFunc = actor.HandlerFunc

// ActorRPCResp Actor RPC响应类型映射.
type ActorRPCResp = actor.ActorRPCResp

// ActorAsyncRPCCallback Actor异步RPC回调类型映射.
type ActorAsyncRPCCallback = actor.ActorAsyncRPCCallback

// ActorTimerArgs Actor定时器参数类型映射.
type ActorTimerArgs = actor.ActorTimerArgs

// ActorTimerFunc Actor定时器函数类型映射.
type ActorTimerFunc = actor.ActorTimerFunc

// ActorFuncArgs Actor函数参数类型映射.
type ActorFuncArgs = actor.ActorFuncArgs

// ActorFunc Actor函数类型映射.
type ActorFunc = actor.ActorFunc

// ContextRPCResp 上下文RPC响应类型映射.
type ContextRPCResp = actor.ContextRPCResp

// ContextAsyncRPCCallback 上下文异步RPC回调类型映射.
type ContextAsyncRPCCallback = actor.ContextAsyncRPCCallback

// ContextFuncArgs 上下文函数参数类型映射.
type ContextFuncArgs = actor.ContextFuncArgs

// ContextFunc 上下文函数类型映射.
type ContextFunc = actor.ContextFunc

// ActorAsyncCaller Actor 异步函数调用器
type ActorAsyncCaller = actor.ActorAsyncCaller

// ContextSuspender 上下文挂起器类型映射.
type ContextSuspender = actor.ContextSuspender
