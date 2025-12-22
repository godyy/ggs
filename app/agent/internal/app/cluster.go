package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/godyy/ggs/app/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/cluster"
	"github.com/godyy/ggs/internal/utils"
)

// startCluster 启动集群.
func (a *app) startCluster() error {
	// 获取IP地址
	ip := utils.ResolveLocalIPv4()

	// 创建cluster
	port := a.config.Cluster.Port
	if port == 0 {
		return errors.New("cluster port not specified")
	}
	node := cluster.NewNode(consts.NodeAgent, a.config.Cluster.NodeName, fmt.Sprintf("%s:%d", ip, port))
	clusterCfg := &cluster.ServiceConfig{
		Core:           &a.config.Cluster.Core,
		Self:           node,
		CenterListener: a,
		Handler:        a,
		Logger:         logger.Get(),
	}
	if Env().Debug() {
		clusterCfg.DefCtxTimeout = time.Hour * 1
	}
	cluster, err := cluster.NewService(clusterCfg)
	if err != nil {
		return err
	}

	// 启动cluster
	if err := cluster.Start(); err != nil {
		return err
	}

	a.cluster = cluster
	return nil
}

// stopCluster 停止集群.
func (a *app) stopCluster() {
	a.cluster.Stop()
}

// OnNodeBytes 处理节点字节数据.
// 当节点字节数据到达时，会调用此方法.
func (a *app) OnNodeBytes(remoteNodeId string, data []byte) error {
	return a.actorClient.HandlePacket(remoteNodeId, data)
}

// OnNodeEvents 处理节点变更事件.
func (a *app) OnNodeEvents(events []cluster.NodeEvent) {
	a.nodeSelector.UpdateEvents(events)
}

// OnNodesSync 处理节点全量同步事件.
func (a *app) OnNodesSync(nodes []*cluster.Node) {
	a.nodeSelector.SetNodes(nodes, true)
}
