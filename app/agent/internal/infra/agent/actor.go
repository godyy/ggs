package agent

import (
	"context"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/agent/internal/app"
	"github.com/godyy/ggs/internal/infra/actor"
)

const (
	actorRequestTimeout  = 5 * time.Second
	playerPreRegisterTTL = int64(30)
)

func checkLocation(location gactor.ActorLocation) bool {
	return location.NodeId != "" && (location.ExpireAt <= 0 || location.ExpireAt > time.Now().Unix())
}

// updatePlayerLocation 更新玩家 Actor 位置信息.
func updatePlayerLocation(playerId int64, nodeId string) error {
	_, err := app.ActorRegistry().RegisterActor(gactor.ActorRegisterParams{
		UID: gactor.ActorUID{
			Category: actor.CategoryPlayer.ActorCategory(),
			ID:       playerId,
		},
		NodeId:  nodeId,
		LeaseId: app.ActorRegistry().MakeLeaseID(),
		TTL:     playerPreRegisterTTL,
	})
	return err
}

// getPlayerLocation 获取玩家 Actor 位置信息.
func getPlayerLocation(playerId int64) (gactor.ActorLocation, error) {
	return app.ActorRegistry().GetActorLocation(gactor.ActorUID{
		Category: actor.CategoryPlayer.ActorCategory(),
		ID:       playerId,
	})
}

// pickReadyWorker 从全局 serving worker 池中选择节点。
func pickReadyWorker(_ int64, playerID int64) (string, error) {
	nodeIDs := app.NodeSelector().PickWorker(playerID, 10)
	for _, nodeID := range nodeIDs {
		readiness, err := actor.GetWorkerReadiness(context.Background(), app.RedisClient(), nodeID)
		if err == nil && readiness.State == actor.WorkerStateServing && readiness.NodeID == nodeID {
			return nodeID, nil
		}
	}
	return "", gactor.ErrNoAvailableNode
}

// genSessionId 生成session id.
func genSessionId() uint32 {
	return app.ActorClient().GenSessionId()
}

// connect2Player 连接到指定玩家.
func connect2Player(playerId int64, sessionId uint32) error {
	return app.ActorClient().Connect(playerId, sessionId)
}

// disconnectPlayer 断开与指定玩家的连接.
func disconnectPlayer(playerId int64, sessionId uint32) error {
	return app.ActorClient().Disconnect(playerId, sessionId)
}

// forwardPacket2Player 向指定玩家转发数据包.
func forwardPacket2Player(playerId int64, sessionId uint32, p []byte) error {
	return app.ActorClient().SendRequest(gactor.ClientRequest{
		ID:      playerId,
		SID:     sessionId,
		Timeout: actorRequestTimeout,
		Payload: p,
	})
}
