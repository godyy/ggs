package internal

import (
	"github.com/godyy/gactor"
	pb_c2s "github.com/godyy/ggs/internal/proto/pb/c2s"
)

// App 内部App接口.
type App interface {
	// GenSessionId 生成用于 agent 与 actor 之间建立通信的会话ID.
	GenSessionId() uint32

	// AddAgent 添加Agent.
	AddAgent(a Agent)

	// DelAgent 删除Agent.
	DelAgent(a Agent)

	// GetAgent 获取Agent.
	GetAgent(playerId int64) Agent

	// Connect2Player 连接到指定玩家.
	Connect2Player(playerId int64, sessionId uint32) error

	// DisconnectPlayer 断开与指定玩家的连接.
	DisconnectPlayer(playerId int64, sessionId uint32) error

	// ForwardPacket2Player 向指定玩家转发数据包.
	ForwardPacket2Player(playerId int64, sessionId uint32, p []byte) error
}

// Agent 内部 Agent 接口.
type Agent interface {
	// PlayerId Agent 关联的 PlayerId.
	PlayerId() int64

	// SessionId Agent 与 Player 建立会话使用的 SessionId.
	SessionId() uint32

	// ReceivePacket 接收上游数据包.
	ReceivePacket(p gactor.Buffer) error

	// Stop Agent 停机.
	Stop(reason pb_c2s.DisconnectPush_Reason)
}
