package actor

import (
	"fmt"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/base/nodeutil"
	"github.com/godyy/ggs/internal/infra/actor/protocol/registry/c2s"
	"github.com/godyy/ggs/internal/infra/actor/protocol/registry/s2s"
	"github.com/godyy/ggskit/infra/actor"
	"github.com/godyy/ggskit/infra/cluster"
	"go.uber.org/zap"
)

// ProtoRegistry Actor协议注册表.
var ProtoRegistry = &actor.ProtoRegistry{
	C2S: c2s.Registry,
	S2S: s2s.Registry,
}

// NewCodec 创建Actor编码器.
func NewCodec() *Codec {
	codec, _ := actor.NewCodec(&actor.CodecConfig{
		ProtoRegistry: ProtoRegistry,
	})
	return codec
}

// NewRegistry 创建Actor注册表函数映射.
var NewRegistry = actor.NewRegistry

// NewServerStore 创建Actor服务器存储函数映射.
var NewServerStore = actor.NewServerStore

// RouterConfig Actor路由配置.
type RouterConfig struct {
	ServerStore *ServerStore
}

// NewRouter 创建Actor路由.
func NewRouter(cfg RouterConfig) *Router {
	serverStore := cfg.ServerStore
	router, _ := actor.NewRouter(actor.RouterConfig{
		NodeGroup:      getNodeGroup,
		ActorFixedNode: getActorFixedNode,
		ActorNodeGroup: func(uid actor.ActorUID) (group string, ok bool) {
			return getActorNodeGroup(uid, serverStore)
		},
	})
	return router
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
	switch Category(uid.Category) {
	case CategoryServer:
		return cluster.MakeNodeID(consts.NodeGame, nodeutil.MakeServerNodeName(uid.ID)), true
	default:
		return "", false
	}
}

// getActorNodeGroup 获取Actor节点分组.
func getActorNodeGroup(uid gactor.ActorUID, serverStore *ServerStore) (string, bool) {
	switch Category(uid.Category) {
	case CategoryPlayer:
		serverID, ok := getActorServerID(uid, serverStore)
		if !ok {
			return "", false
		}
		return makeNodeGroup(consts.NodeGame, serverID), true
	default:
		return "", false
	}
}

// getActorServerID 获取 Actor 所属服务器ID.
func getActorServerID(uid gactor.ActorUID, serverStore *ServerStore) (int64, bool) {
	switch Category(uid.Category) {
	case CategoryServer:
		return uid.ID, true
	case CategoryPlayer:
		serverID, ok, err := serverStore.GetActorServer(uid)
		if err != nil || !ok || serverID <= 0 {
			// 打印错误日志，方便排查获取玩家Actor所属服务器ID失败的问题
			if err != nil {
				logger.Get().Error("get actor server failed",
					zap.String("category", Category(uid.Category).String()),
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
