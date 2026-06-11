package config

import (
	"github.com/godyy/ggskit/base/config"
	"github.com/godyy/ggskit/base/db/redis"
	"github.com/godyy/ggskit/base/logger"
	"github.com/godyy/ggskit/infra/cluster"
	pkgerrors "github.com/pkg/errors"
)

// Config 配置.
type Config struct {
	// Port 服务端口.
	Port int

	// Cluster 集群配置.
	Cluster struct {
		// Port 集群端口.
		Port int

		// Core 核心配置.
		Core cluster.Config
	}

	// DB 数据库配置.
	DB struct {
		// Redis 配置.
		Redis *redis.Config
	}

	// TokenKeyPath 令牌密钥文件路径.
	TokenKeyPath string

	// HttpPort HTTP端口.
	HttpPort int

	// EnablePProf 是否启用pprof.
	EnablePProf bool

	// Log 日志配置
	Log *logger.Config
}

// Load 从指定路径加载配置文件.
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	if err := config.LoadFile(cfg, configPath); err != nil {
		return nil, pkgerrors.WithMessage(err, "load file")
	}

	return cfg, nil
}
