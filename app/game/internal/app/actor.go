package app

import (
	"context"
	"time"

	actors2 "github.com/godyy/ggs/app/game/internal/actors"
	"github.com/godyy/ggs/app/game/internal/env"

	actor2 "github.com/godyy/ggs/internal/core/actor"

	"github.com/godyy/gactor"
	hplayer "github.com/godyy/ggs/app/game/internal/handlers/player"
	hserver "github.com/godyy/ggs/app/game/internal/handlers/server"
	mactor "github.com/godyy/ggs/app/game/internal/modules/actor"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/logger"
	pkgerrors "github.com/pkg/errors"
)

// actorDefineList Actor 定义列表
var actorDefineList = []gactor.IActorDefine{
	&gactor.ActorDefine{
		ActorDefineCommon: &gactor.ActorDefineCommon{
			Name:                       actor2.CategoryName(actor2.CategoryServer),
			Category:                   actor2.CategoryServer,
			Priority:                   0,
			MessageBoxSize:             1000,
			MaxTriggeredTimerAmount:    10,
			MaxCompletedAsyncRPCAmount: 10,
			RecycleTime:                0, // 不回收
			Handler:                    hserver.Handle,
		},
		BehaviorCreator: func(a gactor.Actor) gactor.ActorBehavior {
			return actors2.NewServer(a)
		},
	},
	&gactor.CActorDefine{
		ActorDefineCommon: &gactor.ActorDefineCommon{
			Name:                       actor2.CategoryName(actor2.CategoryPlayer),
			Category:                   actor2.CategoryPlayer,
			Priority:                   99,
			MessageBoxSize:             10,
			MaxTriggeredTimerAmount:    10,
			MaxCompletedAsyncRPCAmount: 10,
			RecycleTime:                time.Minute * 30,
			Handler:                    hplayer.Handle,
		},
		BehaviorCreator: func(c gactor.CActor) gactor.CActorBehavior {
			return actors2.NewPlayer(c)
		},
	},
}

func (a *app) startActor() error {
	// 创建meta数据驱动.
	a.actorMetaDriver = actor2.NewMetaDriver(redis.Inst())

	// 启动actor模块
	if service, err := mactor.Start(actorDefineList, a); err != nil {
		return err
	} else {
		a.actorService = service
	}

	// 启动actor相关服务.
	if err := a.startActorServices(); err != nil {
		return err
	}

	return nil
}

// startActorServices 启动Actor相关服务.
func (a *app) startActorServices() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	// 启动Server
	if err := mactor.StartActor(ctx, gactor.ActorUID{
		Category: actor2.CategoryServer,
		ID:       env.All().ServerID(),
	}); err != nil {
		return pkgerrors.WithMessage(err, "start server actor")
	}

	return nil
}

func (a *app) stopActor() {
	if err := mactor.Stop(); err != nil {
		logger.GetLogger().Errorf("stop actor service failed, %v", err)
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
	return &a.Codec
}

// GetTimeSystem 获取时间系统.
func (a *app) GetTimeSystem() gactor.TimeSystem {
	return gactor.DefTimeSystem
}

// GetMonitor 获取监控器.
func (a *app) GetMonitor() gactor.ServiceMonitor {
	return nil
}

// NodeId 返回本地节点ID.
func (a *app) NodeId() string {
	return a.clusterAgent.NodeId()
}

// Send 发送字节数据 b 到 nodeId 指定的节点.
func (a *app) Send(ctx context.Context, nodeId string, b []byte) error {
	return a.clusterAgent.Send2Node(ctx, nodeId, b)
}
