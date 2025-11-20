package app

import (
	"time"

	"github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/ggs/internal/modules/cluster"
	"github.com/godyy/ggs/internal/utils"
)

// startCluster 启动集群.
func (a *app) startCluster() error {
	// 获取IP地址
	ip := utils.ResolveLocalIPv4()

	// 创建cluster
	clusterCfg := &cluster.ServiceConfig{
		Core:    &a.config.Cluster,
		IP:      ip,
		Handler: a,
		Logger:  logger.GetLogger(),
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
