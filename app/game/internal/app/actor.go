package app

import (
	"context"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/internal/base/consts"
	"github.com/godyy/ggs/app/internal/infra/actors"
	actorsdefine "github.com/godyy/ggs/app/internal/infra/actors/define"
	"github.com/godyy/ggs/app/internal/infra/actors/persist"
	rediscli "github.com/godyy/ggs/internal/base/db/redis/cli"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/cluster"
	pbs2s "github.com/godyy/ggs/internal/proto/pb/s2s"
	"github.com/godyy/gtimewheel"
	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap"
)

// ActorService 返回 Actor 服务.
func ActorService() *actor.Service {
	return appInst.actorService
}

func (a *app) startActor() error {
	// 初始化 actors.
	actors.Init(&actors.InitConfig{
		Persist:           &persist.InitConfig{BD: a.mongobd},
		DB:                a.env.DB(),
		AsyncSaveCallback: a.actorAsyncSaveCallback,
	})

	// 创建meta数据驱动.
	a.actorMetaDriver = actor.NewMetaDriver(rediscli.Get())

	// 创建Actor服务.
	actorConfig := &actor.ServiceConfig{
		Core: &gactor.ServiceConfig{
			NodeId: cluster.MakeNodeID(consts.NodeGame, a.config.Cluster.NodeName),
			ActorConfig: gactor.ActorConfig{
				ActorDefines:        actorsdefine.GetDefineList(),
				ClientActorCategory: actors.CategoryPlayer.ActorCategory(),
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
		Logger: logger.Get(),
	}
	if a.env.Debug() {
		actorConfig.Core.DefCtxTimeout = time.Hour * 1
		actorConfig.Core.DefRPCTimeout = time.Hour * 1
	}
	a.actorService = actor.NewService(actorConfig)

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
		Category: actors.CategoryServer.ActorCategory(),
		ID:       a.env.ServerID(),
	}); err != nil {
		return pkgerrors.WithMessage(err, "start server actor")
	}

	return nil
}

func (a *app) stopActor() {
	a.actorService.Stop()
}

func (a *app) actorAsyncSaveCallback(uid gactor.ActorUID, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ActorCastTimeout)
	defer cancel()
	if err := a.actorService.Cast(ctx, uid, &pbs2s.ActorSaveResult{
		Success: err == nil,
	}); err != nil {
		logger.Get().ErrorFields("cast persist result to actor",
			zap.String("category", actors.Category(uid.Category).String()),
			zap.Int64("id", uid.ID),
			zap.NamedError("error", err),
		)
	}
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
