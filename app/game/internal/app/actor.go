package app

import (
	"context"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/actor"
	actordefine "github.com/godyy/ggs/internal/base/actor/define"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/logger"
	mactor "github.com/godyy/ggs/internal/modules/actor"
	"github.com/godyy/gtimewheel"
	pkgerrors "github.com/pkg/errors"
)

// ActorService 返回 Actor 服务.
func ActorService() *mactor.Service {
	return appInst.actorService
}

func (a *app) startActor() error {
	// 创建meta数据驱动.
	a.actorMetaDriver = mactor.NewMetaDriver(redis.Inst())

	// 创建Actor服务.
	actorConfig := &mactor.ServiceConfig{
		Core: &gactor.ServiceConfig{
			NodeId: a.config.Cluster.NodeId,
			ActorConfig: gactor.ActorConfig{
				ActorDefines:        actordefine.GetDefineList(),
				ClientActorCategory: actor.CategoryPlayer.Uint16(),
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
			Handler: a,
		},
		Logger: logger.GetLogger(),
	}
	if a.env.Debug() {
		actorConfig.Core.DefCtxTimeout = time.Hour * 1
		actorConfig.Core.DefRPCTimeout = time.Hour * 1
	}
	a.actorService = mactor.NewService(actorConfig)

	// 启动actor服务
	if err := a.actorService.Start(); err != nil {
		return err
	}

	// 启动全局Actor.
	if err := a.startGlobalActors(); err != nil {
		return err
	}

	return nil
}

// startGlobalActors 启动全局Actor.
func (a *app) startGlobalActors() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	// 启动Server
	if err := a.actorService.StartActor(ctx, gactor.ActorUID{
		Category: actor.CategoryServer.Uint16(),
		ID:       a.env.ServerID(),
	}); err != nil {
		return pkgerrors.WithMessage(err, "start server actor")
	}

	return nil
}

func (a *app) stopActor() {
	a.actorService.Stop()
}

// GetMetaDriver 获取 Meta 数据驱动.
func (a *app) GetMetaDriver() gactor.MetaDriver {
	return a.actorMetaDriver
}

// GetNetAgent 获取网络代理.
func (a *app) GetNetAgent() gactor.NetAgent {
	return a
}

// GetPacketCodec 获取数据包编解码器.
func (a *app) GetPacketCodec() gactor.PacketCodec {
	return &a.actorCodec
}

// GetTimeSystem 获取时间系统.
func (a *app) GetTimeSystem() gactor.TimeSystem {
	return gactor.DefTimeSystem
}

// GetMonitor 获取监控器.
func (a *app) GetMonitor() gactor.ServiceMonitor {
	return nil
}

// Send2Node 发送字节数据 b 到 nodeId 指定的节点.
func (a *app) Send2Node(ctx context.Context, nodeId string, b []byte) error {
	return a.cluster.Send2Node(ctx, nodeId, b)
}
