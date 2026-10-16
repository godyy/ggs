package nodeutil

import (
	"errors"
	"strconv"

	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggskit/infra/cluster"
)

// MakeServerNodeName 使用 serverId 构造节点名。
func MakeServerNodeName(serverId int64) string {
	return strconv.FormatInt(serverId, 10)
}

// NewServerNode 使用 category、serverId 和 addr 构造服务自身节点。
func NewServerNode(category string, serverId int64, addr string) *cluster.Node {
	return &cluster.Node{
		Category: category,
		Name:     MakeServerNodeName(serverId),
		Addr:     addr,
		ServerId: serverId,
	}
}

// NewWorkerNode 构造全局共享的可迁移 Actor worker 节点。
func NewWorkerNode(workerID, addr string) (*cluster.Node, error) {
	if workerID == "" {
		return nil, errors.New("worker id is required")
	}
	return &cluster.Node{
		Category: consts.NodeGameWorker,
		Name:     workerID,
		Addr:     addr,
	}, nil
}
