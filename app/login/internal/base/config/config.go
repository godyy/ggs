package config

import (
	"github.com/godyy/ggs/internal/libs/config"
	"github.com/godyy/ggs/internal/libs/db/mongo"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/logger"
	pkgerrors "github.com/pkg/errors"
)

// Config 配置.
type Config struct {
	// Port 服务端口.
	Port int

	// AuthKeyPath 鉴权密钥文件路径.
	AuthKeyPath string

	// SignKeyPath 签名密钥文件路径.
	SignKeyPath string

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

// Load 从指定路径加载配置文件.
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	if err := config.LoadFile(cfg, configPath); err != nil {
		return nil, pkgerrors.WithMessage(err, "load file")
	}

	return cfg, nil
}
