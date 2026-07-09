package app

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/nodeutil"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/service"
	"github.com/godyy/ggskit/infra/cluster"
	pkgerrors "github.com/pkg/errors"
)

func (a *app) startActor() error {
	selfNodeId := cluster.MakeNodeID(consts.NodePlatform, nodeutil.MakeServerNodeName(Env().ServerID()))

	// 创建注册表.
	var err error
	a.actorRegistry, err = actor.NewRegistry(a.redisClient)
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor registry")
	}
	a.actorServerStore, err = actor.NewServerStore(a.redisClient)
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor server store")
	}
	a.actorRouter = actor.NewRouter(actor.RouterConfig{
		ServerStore: a.actorServerStore,
	})
	a.actorRouter.SetNodes([]*cluster.Node{
		selfNode(""),
	}, true)

	// 创建Actor服务.
	a.actorService, err = service.NewOneWayService(&service.ServiceConfig{
		NodeId:  selfNodeId,
		Handler: a,
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

	return nil
}

func (a *app) stopActor() {
	a.actorService.Stop()
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

func ActorRegistry() gactor.ActorRegistry {
	return appInst.actorRegistry
}

// ActorService 获取 Actor 服务.
func ActorService() *actor.Service {
	return appInst.actorService
}
