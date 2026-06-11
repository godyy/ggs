package internal

import (
	"net"
	"sync"

	"github.com/godyy/gactor"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
)

// Agent 内部 Agent 接口.
type Agent interface {
	// PlayerId Agent 关联的 PlayerId.
	PlayerId() int64

	// SessionId Agent 与 Player 建立会话使用的 SessionId.
	SessionId() uint32

	// ReceivePacket 接收上游数据包.
	ReceivePacket(p gactor.Buffer) error

	// Stop Agent 停机.
	Stop(reason pbc2s.DisconnectPush_Reason)
}

// StartAgent 启动Agent.
var StartAgent func(conn net.Conn, sessionKey []byte, readInsideIndependentRoutine bool) error

var agents sync.Map

// AddAgent 添加 Agent.
func AddAgent(a Agent) {
	agents.Store(a.PlayerId(), a)
}

// DelAgent 删除 Agent.
func DelAgent(a Agent) {
	agents.Delete(a.PlayerId())
}

// GetAgent 获取 Agent.
func GetAgent(playerId int64) Agent {
	a, ok := agents.Load(playerId)
	if !ok {
		return nil
	}
	return a.(Agent)
}

// GetAgentBySessionId 根据 playerId 和 sessionId 获取 Agent.
func GetAgentBySessionId(playerId int64, sessionId uint32) Agent {
	if agent := GetAgent(playerId); agent != nil && agent.SessionId() == sessionId {
		return agent
	} else {
		return nil
	}
}

// StopAllAgents 停止所有 Agent.
func StopAllAgents() {
	agents.Range(func(key, value interface{}) bool {
		agent := value.(Agent)
		agent.Stop(pbc2s.DisconnectPush_SystemError)
		return true
	})
}
