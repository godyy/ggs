package app

import (
	"net"
	"sync"

	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"

	"github.com/godyy/ggs/internal/core/actor"
	"github.com/godyy/ggs/internal/core/cluster"
	"github.com/godyy/ggs/internal/core/crypto"

	"github.com/godyy/gactor"
	"github.com/godyy/gcluster"
	"github.com/godyy/ggs/app/agent/internal/app/internal"
	icrypto "github.com/godyy/ggs/app/internal/crypto"
	pkgerrors "github.com/pkg/errors"
)

var appInst *app

// Start 启动应用.
func Start() error {
	appInst = &app{}

	// 初始化加密工具
	if err := appInst.initCrypto(); err != nil {
		return pkgerrors.WithMessage(err, "init crypto")
	}

	// 启动 Actor 服务.
	if err := appInst.startActor(); err != nil {
		return pkgerrors.WithMessage(err, "start actor")
	}

	// 启动 cluster.
	if err := appInst.startCluster(); err != nil {
		return pkgerrors.WithMessage(err, "start cluster")
	}

	// 启动对 c 端监听服务.
	if err := appInst.startListen(); err != nil {
		return pkgerrors.WithMessage(err, "start listening")
	}

	return nil
}

// Stop 停止应用.
func Stop() {
	// 停止对 c 端监听服务.
	appInst.stopListen()

	// 停止所有 agent.
	appInst.stopAllAgents()

	// 停止 Actor 服务.
	appInst.stopActor()

	// 停止 cluster.
	appInst.stopCluster()
}

type app struct {
	// 对接 c 端
	listener net.Listener

	// cluster.
	clusterCenter *cluster.Center
	clusterAgent  *gcluster.Agent

	// actor.
	actorMetaDriver *actor.MetaDriver
	actorClient     *gactor.Client

	// agent.
	agents sync.Map

	// crypto.
	secretDecryptor crypto.Decryptor
}

// initCrypto 初始化加密工具.
func (a *app) initCrypto() error {
	if secretDecryptor, err := icrypto.CreateRSADecryptor(); err != nil {
		return pkgerrors.WithMessage(err, "create secret decryptor")
	} else {
		a.secretDecryptor = secretDecryptor
	}
	return nil
}

// AddAgent 添加 agent.
func (a *app) AddAgent(agent internal.Agent) {
	a.agents.Store(agent.PlayerId(), agent)
}

// DelAgent 删除 agent.
func (a *app) DelAgent(agent internal.Agent) {
	a.agents.CompareAndDelete(agent.PlayerId(), agent)
}

// GetAgent 获取 agent.
func (a *app) GetAgent(playerId int64) internal.Agent {
	if value, loaded := a.agents.Load(playerId); loaded {
		return value.(internal.Agent)
	} else {
		return nil
	}
}

// getAgentBySessionId 根据 playerId 和 sessionId 获取 agent.
func (a *app) getAgentBySessionId(playerId int64, sessionId uint32) internal.Agent {
	if agent := a.GetAgent(playerId); agent != nil && agent.SessionId() == sessionId {
		return agent
	} else {
		return nil
	}
}

// stopAllAgents 停止所有 agent.
func (a *app) stopAllAgents() {
	a.agents.Range(func(key, value interface{}) bool {
		agent := value.(internal.Agent)
		agent.Stop(pbc2s.DisconnectPush_SystemError)
		return true
	})
}
