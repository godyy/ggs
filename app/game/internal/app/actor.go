package app

import (
	"context"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/handler"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/base/nodeutil"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/persist"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
	"github.com/godyy/ggs/internal/infra/actor/service"
	"github.com/godyy/ggskit/infra/cluster"
	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap"
)

// ActorService 返回 Actor 服务.
func ActorService() *actor.Service {
	return appInst.actorService
}

func (a *app) startActor() error {
	selfNodeId := cluster.MakeNodeID(consts.NodeGame, nodeutil.MakeServerNodeName(Env().ServerID()))

	// 初始化 actors.
	actor.Init(&actor.InitConfig{
		Persist:           &persist.InitConfig{BD: a.mongobd},
		DB:                a.env.DB(),
		AsyncSaveCallback: a.actorAsyncSaveCallback,
	})

	// 创建注册表.
	var err error
	a.actorRegistry, err = actor.CreateRegistry(&actor.RegistryConfig{
		RedisCli: a.redisClient,
	})
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor registry")
	}
	a.actorServerStore, err = actor.CreateServerStore(&actor.ServerStoreConfig{
		RedisCli: a.redisClient,
	})
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor server store")
	}
	a.actorRouter = actor.NewRouter(actor.RouterConfig{ServerStore: a.actorServerStore})
	a.actorRouter.SetNodes([]*cluster.Node{
		selfNode(""),
	}, true)

	// 创建Actor服务.
	a.actorService, err = service.NewClientService(&service.ServiceConfig{
		NodeId:         selfNodeId,
		Handler:        a,
		RequestHandler: handler.Handle,
	})

	if err != nil {
		return pkgerrors.WithMessage(err, "new actor service")
	}
	a.actorCodec = actor.NewCodec()
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor codec")
	}

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
		Category: actor.CategoryServer.ActorCategory(),
		ID:       a.env.ServerID(),
	}); err != nil {
		return pkgerrors.WithMessage(err, "start server actor")
	}

	// 启动chatmgr
	if err := a.actorService.StartActor(ctx, gactor.ActorUID{
		Category: actor.CategoryChatMgr.ActorCategory(),
		ID:       a.env.ServerID(),
	}); err != nil {
		return pkgerrors.WithMessage(err, "start chat mgr actor")
	}
	return nil
}

func (a *app) stopActor() {
	a.actorService.Stop()
}

func (a *app) actorAsyncSaveCallback(uid gactor.ActorUID, err error) {
	if castErr := a.actorService.Cast(uid, &pbs2s.ActorSaveResultNtf{
		Success: err == nil,
	}); castErr != nil {
		logger.Get().ErrorFields("cast persist result to actor",
			zap.String("category", actor.Category(uid.Category).String()),
			zap.Int64("id", uid.ID),
			zap.NamedError("error", castErr),
		)
	}
}

// GetActorRegistry 获取 Actor 注册表.
func (a *app) GetActorRegistry() gactor.ActorRegistry {
	return a.actorRegistry
}

// GetActorRouter 获取 Actor 路由.
func (a *app) GetActorRouter() gactor.ActorRouter {
	return a.actorRouter
}

// GetNetAgent 获取网络代理.
func (a *app) GetNetAgent() gactor.NetAgent {
	return a
}

// GetPacketCodec 获取数据包编解码器.
func (a *app) GetPacketCodec() gactor.PacketCodec {
	return a.actorCodec
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
func (a *app) Send2Node(nodeId string, b []byte) error {
	return a.cluster.Send2Node(nodeId, b)
}
