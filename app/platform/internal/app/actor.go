package app

import (
	"fmt"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/base/nodeutil"
	iactor "github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/protocol/registry/s2s"
	"github.com/godyy/ggskit/infra/actor"
	"github.com/godyy/ggskit/infra/cluster"
	"github.com/godyy/gtimewheel"
	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap"
)

func (a *app) startActor() error {
	selfNodeId := cluster.MakeNodeID(consts.NodePlatform, nodeutil.MakeServerNodeName(Env().ServerID()))

	a.actorProtoReg = &actor.ProtoRegistry{S2S: s2s.Registry}

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
	a.actorRouter, err = actor.NewRouter(actor.RouterConfig{
		NodeGroup:      getNodeGroup,
		ActorFixedNode: getActorFixedNode,
		ActorNodeGroup: getActorNodeGroup,
	})
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor router")
	}
	a.actorRouter.SetNodes([]*cluster.Node{
		selfNode(""),
	}, true)

	// 创建Actor服务.
	actorConfig := &actor.ServiceConfig{
		Core: &gactor.ServiceConfig{
			NodeId: selfNodeId,
			ActorConfig: gactor.ActorConfig{
				ActorDefines: actors.GetDefineList(),
				Handler:      func(ctx *gactor.Context) {},
			},
			TimerConfig: gactor.TimerConfig{
				TimeWheelLevels: []gtimewheel.LevelConfig{
					{Name: "second", Span: 50 * time.Millisecond, Slots: 20},
					{Name: "minute", Span: 1 * time.Second, Slots: 60},
					{Name: "hour", Span: 1 * time.Minute, Slots: 60},
					{Name: "day", Span: 1 * time.Hour, Slots: 24},
					{Name: "month", Span: 1 * time.Hour * 24, Slots: 30},
				},
				MaxTimerDelay:  time.Hour * 24 * 7,
				MaxTimerAmount: 1000,
			},
			RPCConfig: gactor.RPCConfig{
				MaxRPCCallAmount: 1000,
			},
			MaxRTT:  50,
			Handler: a,
		},
		Logger:        logger.Get(),
		ProtoRegistry: a.actorProtoReg,
	}
	if Env().Debug() {
		actorConfig.Core.DefRPCTimeout = time.Hour * 1
	}
	a.actorService, err = actor.NewService(actorConfig)
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor service")
	}
	a.actorCodec, err = actor.NewCodec(&actor.CodecConfig{
		ProtoRegistry: a.actorProtoReg,
	})
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

// getNodeGroup 获取节点分组.
func getNodeGroup(node *cluster.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	return makeNodeGroup(node.Category, node.ServerId), true
}

// getActorFixedNode 获取 Actor 固定节点.
func getActorFixedNode(uid gactor.ActorUID) (string, bool) {
	switch iactor.Category(uid.Category) {
	case iactor.CategoryServer:
		return cluster.MakeNodeID(consts.NodeGame, nodeutil.MakeServerNodeName(uid.ID)), true
	default:
		return "", false
	}
}

// getActorNodeGroup 获取Actor节点分组.
func getActorNodeGroup(uid gactor.ActorUID) (string, bool) {
	switch iactor.Category(uid.Category) {
	case iactor.CategoryPlayer:
		serverID, ok := getActorServerID(uid)
		if !ok {
			return "", false
		}
		return makeNodeGroup(consts.NodeGame, serverID), true
	default:
		return "", false
	}
}

// getActorServerID 获取 Actor 所属服务器ID.
func getActorServerID(uid gactor.ActorUID) (int64, bool) {
	switch iactor.Category(uid.Category) {
	case iactor.CategoryServer:
		return uid.ID, true
	case iactor.CategoryPlayer:
		serverID, ok, err := appInst.actorServerStore.GetActorServer(uid)
		if err != nil || !ok || serverID <= 0 {
			// 打印错误日志，方便排查获取玩家Actor所属服务器ID失败的问题
			if err != nil {
				logger.Get().Error("get actor server failed",
					zap.String("category", iactor.Category(uid.Category).String()),
					zap.Int64("actorId", uid.ID),
					zap.Error(err))
			}
			return 0, false
		}
		return serverID, true
	default:
		return 0, false
	}
}

// makeNodeGroup 生成节点分组.
func makeNodeGroup(category string, serverId int64) string {
	if category == consts.NodeGame {
		return fmt.Sprintf("%s/%d", consts.NodeGame, serverId)
	}
	return category
}

// ActorService 获取 Actor 服务.
func ActorService() *actor.Service {
	return appInst.actorService
}
