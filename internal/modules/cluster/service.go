package cluster

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/godyy/gcluster"
	clusternet "github.com/godyy/gcluster/net"
	"github.com/godyy/glog"
)

// ServiceConfig 集群服务配置.
type ServiceConfig struct {
	// Core 核心配置.
	Core *Config

	// IP 节点IP地址.
	IP string

	// Handler 集群代理处理函数.
	Handler gcluster.AgentHandler

	// Logger 日志记录器.
	Logger glog.Logger

	// DefCtxTimeout 默认上下文超时时间.
	DefCtxTimeout time.Duration
}

// Service 集群服务.
type Service struct {
	center *Center         // 数据中心.
	agent  *gcluster.Agent // 集群代理.
}

// NewService 构造集群服务.
func NewService(cfg *ServiceConfig) (*Service, error) {
	// 构造节点地址
	addr := fmt.Sprintf("%s:%d", cfg.IP, cfg.Core.Port)

	// 创建center
	center := NewCenter(&CenterConfig{
		EndPoints: cfg.Core.EtcdEndPoints,
		Root:      cfg.Core.EtcdRoot,
		Self: &Node{
			ID:   cfg.Core.NodeId,
			Addr: addr,
		},
		Log: cfg.Logger.Named("cluster-center"),
	})

	// 创建agent
	agent, err := gcluster.CreateAgent(
		&gcluster.AgentConfig{
			Center: center,
			Net: &clusternet.ServiceConfig{
				NodeId:    cfg.Core.NodeId,
				Addr:      addr,
				Handshake: cfg.Core.Handshake,
				Session:   cfg.Core.Session,
				Dialer: func(addr string) (net.Conn, error) {
					return net.Dial("tcp", addr)
				},
				ListenerCreator: func(addr string) (net.Listener, error) {
					return net.Listen("tcp", addr)
				},
				TimerSystem:                clusternet.NewTimerHeap(),
				ExpectedConcurrentSessions: cfg.Core.ExpectedConcurrentSessions,
				DefCtxTimeout:              cfg.DefCtxTimeout,
			},
			Handler: cfg.Handler,
		},
		gcluster.WithLogger(cfg.Logger),
		gcluster.WithServiceOptions(clusternet.WithServiceLogger(cfg.Logger)),
	)
	if err != nil {
		return nil, err
	}

	return &Service{
		center: center,
		agent:  agent,
	}, nil
}

// Start 启动集群服务.
func (s *Service) Start() error {
	// 启动agent
	if err := s.agent.Start(); err != nil {
		return err
	}

	// 启动center
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := s.center.Start(ctx); err != nil {
		return err
	}

	return nil
}

// Start 关闭集群服务.
func (s *Service) Stop() {
	// 关闭center
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	s.center.Close(ctx)

	// 关闭agent
	s.agent.Close()
}

// Send2Node 发送字节数据 b 到 nodeId 指定的节点.
func (s *Service) Send2Node(ctx context.Context, nodeId string, b []byte) error {
	return s.agent.Send2Node(ctx, nodeId, b)
}
