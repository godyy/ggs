package app

import (
	"context"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/agent/internal"
	"github.com/godyy/ggs/app/agent/internal/base/log"
	"github.com/godyy/ggs/app/internal/base/consts"
	"github.com/godyy/ggs/app/internal/infra/actors"
	rediscli "github.com/godyy/ggs/internal/base/db/redis/cli"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/cluster"
	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
)

func (a *app) startActor() error {
	// 创建meta数据驱动.
	a.actorMetaDriver = actor.NewMetaDriver(rediscli.Get())

	// 创建 actor 客户端.
	clientCfg := &actor.ClientConfig{
		Core: &gactor.ClientConfig{
			NodeId:            cluster.MakeNodeID(consts.NodeAgent, a.config.Cluster.NodeName),
			ActorCategory:     actors.CategoryPlayer.ActorCategory(),
			DefRequestTimeout: time.Second * 10,
			Handler:           a,
		},
		Logger: logger.Get(),
	}
	if Env().Debug() {
		clientCfg.Core.DefCtxTimeout = time.Hour * 1
	}
	a.actorClient = actor.NewClient(clientCfg)

	return nil
}

func (a *app) stopActor() {
	a.actorClient.Stop()
}

// GetMetaDriver 获取 Meta 数据驱动.
func (a *app) GetMetaDriver() gactor.MetaDriver {
	return a.actorMetaDriver
}

// GetNetAgent 获取网络代理.
func (a *app) GetNetAgent() gactor.NetAgent {
	return a
}

// GetBytesManager 获取字节切片管理器.
func (a *app) GetBytesManager() gactor.BytesManager {
	return a
}

// Send2Node 向集群中 nodeId 指向的节点发送字节数据.
func (a *app) Send2Node(ctx context.Context, nodeId string, b []byte) error {
	return a.cluster.Send2Node(ctx, nodeId, b)
}

func (a *app) GetBytes(cap int) []byte {
	return make([]byte, 0, cap)
}

func (a *app) PutBytes(b []byte) {}

// HandleResponse 处理 ClientResponse.
func (a *app) HandleResponse(resp gactor.ClientResponse) {
	agent := internal.GetAgent(resp.ID)
	if agent == nil {
		return
	}

	if resp.Err != nil {
		logger.Get().ErrorFields("handle actor response error", log.FldPlayerId(resp.ID), log.FldError(resp.Err))
		agent.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}

	if err := agent.ReceivePacket(resp.Payload); err != nil {
		logger.Get().ErrorFields("agent receive actor response packet failed", log.FldPlayerId(resp.ID), log.FldError(err))
		agent.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}
}

// HandlePush 处理 ClientPush.
func (a *app) HandlePush(push gactor.ClientPush) {
	agent := internal.GetAgent(push.ID)
	if agent == nil {
		return
	}

	if err := agent.ReceivePacket(push.Payload); err != nil {
		logger.Get().ErrorFields("agent receive actor push packet failed", log.FldPlayerId(push.ID), log.FldError(err))
		agent.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}
}

// HandleDisconnect 处理 Actor 断开连接.
func (a *app) HandleDisconnect(id int64, sid uint32) {
	agent := internal.GetAgent(id)
	if agent == nil {
		return
	}
	agent.Stop(pbc2s.DisconnectPush_SystemError)
}
