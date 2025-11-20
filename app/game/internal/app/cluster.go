package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/godyy/ggs/internal/core/cluster"

	"github.com/godyy/gcluster"
	clusternet "github.com/godyy/gcluster/net"
	"github.com/godyy/ggs/app/game/internal/config"
	"github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/ggs/internal/utils"
)

// initCluster 初始化集群相关.
func (a *app) initCluster() error {
	// 获取集群配置.
	clusterConfig := &config.GetConfig().Cluster

	// 构造集群内服务地址
	ip := utils.ResolveLocalIPv4()
	clusterAddr := fmt.Sprintf("%s:%d", ip, clusterConfig.Port)

	// 创建集群center.
	if err := a.createClusterCenter(clusterAddr); err != nil {
		return err
	}

	// 创建集群agent.
	if err := a.createClusterAgent(clusterAddr); err != nil {
		return err
	}

	return nil
}

func (a *app) createClusterCenter(addr string) error {
	clusterConfig := &config.GetConfig().Cluster

	a.clusterCenter = cluster.NewCenter(&cluster.CenterConfig{
		EndPoints: clusterConfig.EtcdEndPoints,
		Root:      clusterConfig.EtcdRoot,
		Self: &cluster.Node{
			ID:   clusterConfig.NodeId,
			Addr: addr,
		},
		Log: logger.GetLogger().Named("cluster-center"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return a.clusterCenter.Start(ctx)
}

func (a *app) createClusterAgent(addr string) (err error) {
	clusterConfig := &config.GetConfig().Cluster
	a.clusterAgent, err = gcluster.CreateAgent(
		&gcluster.AgentConfig{
			Center: a.clusterCenter,
			Net: &clusternet.ServiceConfig{
				NodeId:    clusterConfig.NodeId,
				Addr:      addr,
				Handshake: clusterConfig.Handshake,
				Session:   clusterConfig.Session,
				Dialer: func(addr string) (net.Conn, error) {
					return net.Dial("tcp", addr)
				},
				ListenerCreator: func(addr string) (net.Listener, error) {
					return net.Listen("tcp", addr)
				},
				TimerSystem:                clusternet.NewTimerHeap(),
				ExpectedConcurrentSessions: clusterConfig.ExpectedConcurrentSessions,
			},
			Handler: a,
		},
		gcluster.WithLogger(logger.GetLogger()),
		gcluster.WithServiceOptions(clusternet.WithServiceLogger(logger.GetLogger())),
	)
	return
}

// startCluster 启动集群相关.
func (a *app) startCluster() error {
	return a.clusterAgent.Start()
}

func (a *app) stopCluster() {
	if err := a.clusterAgent.Close(); err != nil {
		logger.GetLogger().Errorf("close cluster agent failed, %v", err)
	}
	if err := a.clusterCenter.Close(context.Background()); err != nil {
		logger.GetLogger().Errorf("close cluster center failed, %v", err)
	}
}

// OnNodeBytes 处理节点字节数据.
// 当节点字节数据到达时，会调用此方法.
func (a *app) OnNodeBytes(remoteNodeId string, data []byte) error {
	return a.actorService.HandlePacket(remoteNodeId, data)
}
