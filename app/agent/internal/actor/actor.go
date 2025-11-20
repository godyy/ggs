package actor

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

// GenSessionId 生成session id.
func GenSessionId() uint32 {
	return app.ActorClient().GenSessionId()
}

// Connect2Player 连接到指定玩家.
func Connect2Player(playerId int64, sessionId uint32) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), actorRequestTimeout)
	defer cancel()
	return app.ActorClient().Connect(ctx, playerId, sessionId)
}

// DisconnectPlayer 断开与指定玩家的连接.
func DisconnectPlayer(playerId int64, sessionId uint32) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), actorRequestTimeout)
	defer cancel()
	return app.ActorClient().Disconnect(ctx, playerId, sessionId)
}

// ForwardPacket2Player 向指定玩家转发数据包.
func ForwardPacket2Player(playerId int64, sessionId uint32, p []byte) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), actorRequestTimeout)
	defer cancel()
	return app.ActorClient().SendRequest(ctx, gactor.ClientRequest{
		ID:      playerId,
		SID:     sessionId,
		Payload: p,
	})
}
