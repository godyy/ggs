package actor

import (
	"context"
	"fmt"
	"time"

	"github.com/godyy/ggs/internal/core/actor"
	"github.com/godyy/ggs/internal/libs/logger"
	prototypes "github.com/godyy/ggs/internal/proto/types"

	"github.com/godyy/ggs/app/game/internal/codec"
	"github.com/godyy/ggs/app/game/internal/env"

	"github.com/godyy/gactor"
	"github.com/godyy/gtimewheel"
	"google.golang.org/protobuf/proto"
)

var (
	service *gactor.Service
)

// Start 启动actor服务.
func Start(actorDefineList []gactor.IActorDefine, handler gactor.ServiceHandler) (*gactor.Service, error) {
	// 创建actor服务.
	serviceConfig := &gactor.ServiceConfig{
		ActorConfig: gactor.ActorConfig{
			ActorDefines:        actorDefineList,
			ClientActorCategory: actor.CategoryPlayer,
		},
		TimerConfig: gactor.TimerConfig{
			TimeWheelLevels: []gtimewheel.LevelConfig{
				{Name: "second", Span: 50 * time.Millisecond, Slots: 20},
				{Name: "minute", Span: 1 * time.Second, Slots: 60},
				{Name: "hour", Span: 1 * time.Minute, Slots: 60},
				{Name: "day", Span: 1 * time.Hour, Slots: 24},
				{Name: "month", Span: 1 * time.Hour * 24, Slots: 30},
			},
			MaxTimerDelay:          time.Hour * 24 * 7,
			MaxTriggerdTimerAmount: 10000,
		},
		RPCConfig: gactor.RPCConfig{
			MaxRPCCallAmount: 10000,
		},
		MaxRTT:  50,
		Handler: handler,
	}
	if env.All().Debug() {
		serviceConfig.DefCtxTimeout = time.Hour * 1
		serviceConfig.RPCConfig.DefRPCTimeout = time.Hour * 1
	}
	service = gactor.NewService(serviceConfig, gactor.WithServiceLogger(logger.GetLogger()))

	// 启动actor服务.
	if err := service.Start(); err != nil {
		return nil, err
	}

	return service, nil
}

// Stop 停止actor服务.
func Stop() error {
	return service.Stop()
}

// StartActor 启动Actor.
func StartActor(ctx context.Context, uid gactor.ActorUID) error {
	return service.StartActor(ctx, uid)
}

// Cast 发送消息到目标actor.
func Cast(ctx context.Context, to gactor.ActorUID, msg proto.Message) error {
	pid, ok := prototypes.S2S.GetPid(msg)
	if !ok {
		return fmt.Errorf("msg %T not registered", msg)
	}
	payload := codec.S2SPayload{
		PID: pid,
		Msg: msg,
	}
	return service.Cast(ctx, to, &payload)
}
