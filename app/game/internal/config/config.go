package config

import (
	"github.com/godyy/gcluster/net"
	"github.com/godyy/ggs/internal/libs/config"
	"github.com/godyy/ggs/internal/libs/db/mongo"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/logger"
	pkgerrors "github.com/pkg/errors"
)

// Config 配置.
type Config struct {
	// Stage 环境. dev/prod
	Stage string

	// Cluster 集群配置.
	Cluster ClusterConfig

	// DB 数据库配置.
	DB struct {
		// Redis 配置.
		Redis *redis.Config

		// Mongo 配置.
		Mongo *mongo.Config
	}

	// Log 日志配置
	Log *logger.Config
}

// ClusterConfig 集群配置.
type ClusterConfig struct {
	// NodeId 节点ID.
	NodeId string

	// Port 集群端口号.
	Port int

	// EtcdEndPoints etcd 节点地址列表.
	EtcdEndPoints []string

	// EtcdRoot 用于发现其它节点信息的etcd根路径.
	EtcdRoot string

	// Handshake 握手配置.
	Handshake net.HandshakeConfig

	// Session 会话配置.
	Session net.SessionConfig

	// ExpectedConcurrentSessions 预期同时存在的 Session 数量.
	ExpectedConcurrentSessions int
}

var (
	// 全局配置单例
	inst = &Config{}
)

// Init 从指定路径加载配置文件并初始化全局配置单例
func Init(configPath string) error {
	cfg := &Config{}

	if err := config.LoadFile(cfg, configPath); err != nil {
		return pkgerrors.WithMessage(err, "load file")
	}

	inst = cfg
	return nil
}

// GetConfig 获取全局配置单例
func GetConfig() *Config {
	return inst
}
