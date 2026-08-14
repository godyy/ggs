package service

import (
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/logger"
	iactor "github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggskit/base/env"
	"github.com/godyy/ggskit/infra/actor"
	"github.com/godyy/gtimewheel"
)

// timeWheelLevels 时间轮层级配置.
var timeWheelLevels = []gtimewheel.LevelConfig{
	{Name: "second", Span: 50 * time.Millisecond, Slots: 20},
	{Name: "minute", Span: 1 * time.Second, Slots: 60},
	{Name: "hour", Span: 1 * time.Minute, Slots: 60},
	{Name: "day", Span: 1 * time.Hour, Slots: 24},
	{Name: "month", Span: 1 * time.Hour * 24, Slots: 30},
}

// maxRTT 最大RTT.
const maxRTT = 50

type ServiceConfig struct {
	NodeId         string                // 节点 ID.
	Handler        gactor.ServiceHandler // 处理器.
	RequestHandler gactor.HandlerFunc    // 请求处理函数.
}

// NewClientService 创建对接客户端的Actor服务.
func NewClientService(cfg *ServiceConfig) (*iactor.Service, error) {
	// 创建Actor服务.
	actorConfig := &actor.ServiceConfig{
		Core: &gactor.ServiceConfig{
			NodeId: cfg.NodeId,
			ActorConfig: gactor.ActorConfig{
				ActorDefines:        actors.GetDefineList(),
				ClientActorCategory: iactor.CategoryPlayer.ActorCategory(),
				Handler:             cfg.RequestHandler,
			},
			TimerConfig: gactor.TimerConfig{
				TimeWheelLevels: timeWheelLevels,
				MaxTimerAmount:  10000,
			},
			RPCConfig: gactor.RPCConfig{
				MaxRPCCallAmount: 10000,
			},
			MaxRTT:  maxRTT,
			Handler: cfg.Handler,
		},
		Logger:        logger.Get(),
		ProtoRegistry: iactor.ProtoRegistry,
	}
	if env.Get().Debug() {
		actorConfig.Core.DefRPCTimeout = time.Hour * 1
	}
	return actor.NewService(actorConfig)
}

// NewInternalService 创建内部Actor服务.
func NewInternalService(cfg *ServiceConfig) (*iactor.Service, error) {
	// 创建Actor服务.
	actorConfig := &actor.ServiceConfig{
		Core: &gactor.ServiceConfig{
			NodeId: cfg.NodeId,
			ActorConfig: gactor.ActorConfig{
				ActorDefines: actors.GetDefineList(),
				Handler:      cfg.RequestHandler,
			},
			TimerConfig: gactor.TimerConfig{
				TimeWheelLevels: timeWheelLevels,
				MaxTimerAmount:  10000,
			},
			RPCConfig: gactor.RPCConfig{
				MaxRPCCallAmount: 10000,
			},
			MaxRTT:  maxRTT,
			Handler: cfg.Handler,
		},
		Logger:        logger.Get(),
		ProtoRegistry: iactor.ProtoRegistry,
	}
	if env.Get().Debug() {
		actorConfig.Core.DefRPCTimeout = time.Hour * 1
	}
	return actor.NewService(actorConfig)
}

// NewOneWayService 创建单向Actor服务.
// 该类Actor服务只会调用接口与集群那其它Actor服务器通信, 服务本身不会启动任何Actor.
func NewOneWayService(cfg *ServiceConfig) (*iactor.Service, error) {
	// 创建Actor服务.
	actorConfig := &actor.ServiceConfig{
		Core: &gactor.ServiceConfig{
			NodeId: cfg.NodeId,
			ActorConfig: gactor.ActorConfig{
				ActorDefines: actors.GetDefineList(),
				Handler:      func(ctx *gactor.Context) {},
			},
			TimerConfig: gactor.TimerConfig{
				TimeWheelLevels: timeWheelLevels,
				MaxTimerAmount:  100,
			},
			RPCConfig: gactor.RPCConfig{
				MaxRPCCallAmount: 100,
			},
			MaxRTT:  maxRTT,
			Handler: cfg.Handler,
		},
		Logger:        logger.Get(),
		ProtoRegistry: iactor.ProtoRegistry,
	}
	if env.Get().Debug() {
		actorConfig.Core.DefRPCTimeout = time.Hour * 1
	}
	return actor.NewService(actorConfig)
}
