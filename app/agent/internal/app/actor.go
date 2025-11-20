package app

import (
	"context"
	"time"

	"github.com/godyy/ggs/internal/env"
	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
	"github.com/godyy/ggs/internal/utils/ctxutils"

	"github.com/godyy/ggs/internal/core/actor"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/agent/internal/config"
	log "github.com/godyy/ggs/app/agent/internal/log"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/logger"
)

const (
	actorRequestTimeout = 5 * time.Second
)

func (a *app) startActor() error {
	// 创建meta数据驱动.
	a.actorMetaDriver = actor.NewMetaDriver(redis.Inst())

	// 创建 actor 客户端.
	cfg := &gactor.ClientConfig{
		ActorCategory:     actor.CategoryPlayer,
		DefRequestTimeout: time.Second * 10,
		Handler:           a,
	}
	if env.All().Debug() {
		cfg.DefCtxTimeout = time.Hour * 1
	}
	a.actorClient = gactor.NewClient(cfg,
		gactor.WithClientLogger(logger.GetLogger()),
	)

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

func (a *app) NodeId() string {
	return config.GetConfig().Cluster.NodeId
}

func (a *app) Send(ctx context.Context, nodeId string, b []byte) error {
	return a.clusterAgent.Send2Node(ctx, nodeId, b)
}

func (a *app) GetBytes(cap int) []byte {
	return make([]byte, 0, cap)
}

func (a *app) PutBytes(b []byte) {}

// HandleResponse 处理 ClientResponse.
func (a *app) HandleResponse(resp gactor.ClientResponse) {
	agent := a.getAgentBySessionId(resp.ID, resp.SID)
	if agent == nil {
		return
	}

	if resp.Err != nil {
		logger.GetLogger().ErrorFields("handle actor response error", log.FldPlayerId(resp.ID), log.FldError(resp.Err))
		agent.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}

	if err := agent.ReceivePacket(resp.Payload); err != nil {
		logger.GetLogger().ErrorFields("agent receive actor response packet failed", log.FldPlayerId(resp.ID), log.FldError(err))
		agent.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}
}

// HandlePush 处理 ClientPush.
func (a *app) HandlePush(push gactor.ClientPush) {
	agent := a.getAgentBySessionId(push.ID, push.SID)
	if agent == nil {
		return
	}

	if err := agent.ReceivePacket(push.Payload); err != nil {
		logger.GetLogger().ErrorFields("agent receive actor push packet failed", log.FldPlayerId(push.ID), log.FldError(err))
		agent.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}
}

// HandleDisconnect 处理 Actor 断开连接.
func (a *app) HandleDisconnect(id int64, sid uint32) {
	agent := a.getAgentBySessionId(id, sid)
	if agent == nil {
		return
	}
	agent.Stop(pbc2s.DisconnectPush_SystemError)
}

// GenSessionId 生成用于 agent 与 actor 之间建立通信的会话ID.
func (a *app) GenSessionId() uint32 {
	return a.actorClient.GenSessionId()
}

// Connect2Player 连接到指定玩家.
func (a *app) Connect2Player(playerId int64, sessionId uint32) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), actorRequestTimeout)
	defer cancel()
	return a.actorClient.Connect(ctx, playerId, sessionId)
}

// DisconnectPlayer 断开与指定玩家的连接.
func (a *app) DisconnectPlayer(playerId int64, sessionId uint32) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), actorRequestTimeout)
	defer cancel()
	return a.actorClient.Disconnect(ctx, playerId, sessionId)
}

// ForwardPacket2Player 向指定玩家转发数据包.
func (a *app) ForwardPacket2Player(playerId int64, sessionId uint32, p []byte) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), actorRequestTimeout)
	defer cancel()
	return a.actorClient.SendRequest(ctx, gactor.ClientRequest{
		ID:      playerId,
		SID:     sessionId,
		Payload: p,
	})
}
