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

var (
	// 全局配置单例
	inst *Config
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
