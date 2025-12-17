package agent

import (
	"context"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/agent/internal/app"
	"github.com/godyy/ggs/internal/utils/ctxutils"
)

const (
	actorRequestTimeout = 5 * time.Second
)

// genSessionId 生成session id.
func genSessionId() uint32 {
	return app.ActorClient().GenSessionId()
}

// connect2Player 连接到指定玩家.
func connect2Player(playerId int64, sessionId uint32) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), actorRequestTimeout)
	defer cancel()
	return app.ActorClient().Connect(ctx, playerId, sessionId)
}

// disconnectPlayer 断开与指定玩家的连接.
func disconnectPlayer(playerId int64, sessionId uint32) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), actorRequestTimeout)
	defer cancel()
	return app.ActorClient().Disconnect(ctx, playerId, sessionId)
}

// forwardPacket2Player 向指定玩家转发数据包.
func forwardPacket2Player(playerId int64, sessionId uint32, p []byte) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), actorRequestTimeout)
	defer cancel()
	return app.ActorClient().SendRequest(ctx, gactor.ClientRequest{
		ID:      playerId,
		SID:     sessionId,
		Payload: p,
	})
}
