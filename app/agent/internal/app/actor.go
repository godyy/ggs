package app

import (
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/agent/internal"
	"github.com/godyy/ggs/app/agent/internal/base/log"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/base/nodeutil"
	iactor "github.com/godyy/ggs/internal/infra/actor"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	"github.com/godyy/ggskit/infra/actor"
	"github.com/godyy/ggskit/infra/cluster"
	pkgerrors "github.com/pkg/errors"
)

func (a *app) startActor() error {
	// 创建注册表.
	var err error
	a.actorRegistry, err = actor.NewRegistry(a.redisClient)
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor registry")
	}

	// 创建 actor 客户端.
	clientCfg := &actor.ClientConfig{
		Core: &gactor.ClientConfig{
			NodeId:            cluster.MakeNodeID(consts.NodeAgent, nodeutil.MakeServerNodeName(Env().ServerID())),
			ActorCategory:     iactor.CategoryPlayer.ActorCategory(),
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

// GetActorRegistry 获取 Actor 注册表.
func (a *app) GetActorRegistry() gactor.ActorRegistry {
	return a.actorRegistry
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
func (a *app) Send2Node(nodeId string, b []byte) error {
	return a.cluster.Send2Node(nodeId, b)
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

	if resp.ErrCode != gactor.ErrCodeOK {
		logger.Get().ErrorFields("handle actor response error", log.FldPlayerId(resp.ID), log.FldError(resp.ErrCode))
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
