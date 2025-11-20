package cluster

import (
	clusternet "github.com/godyy/gcluster/net"
)

// Config 集群配置.
type Config struct {
	// NodeId 节点ID.
	NodeId string

	// Port 集群端口号.
	Port int

	// EtcdEndPoints etcd 节点地址列表.
	EtcdEndPoints []string

	// EtcdRoot 用于发现其它节点信息的etcd根路径.
	EtcdRoot string

	// Handshake 握手配置.
	Handshake clusternet.HandshakeConfig

	// Session 会话配置.
	Session clusternet.SessionConfig

	// ExpectedConcurrentSessions 预期同时存在的 Session 数量.
	ExpectedConcurrentSessions int
}
