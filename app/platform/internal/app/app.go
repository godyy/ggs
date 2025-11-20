package app

import (
	applifecycle "github.com/godyy/ggs/app/internal/base/lifecycle"
	"github.com/godyy/ggs/app/platform/internal/base/config"
	"github.com/godyy/ggs/internal/base/env"
	"github.com/godyy/ggs/internal/libs/db/mongo"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/flags"
	"github.com/godyy/ggs/internal/libs/logger"
	pkgerrors "github.com/pkg/errors"
)

var (
	cfg *config.Config // 配置
)

// Start 启动.
func Start() {
	// 解析flag
	flags.Parse()
	defer flags.Clear()

	// 加载配置.
	if c, err := config.Load(configPath()); err != nil {
		panic(pkgerrors.WithMessage(err, "load config"))
	} else {
		cfg = c
	}

	// 初始化日志工具.
	logger.Init(cfg.Log)

	// 启动前回调.
	applifecycle.BeforeStart()

	// 初始化 redis.
	if err := redis.Init(cfg.DB.Redis); err != nil {
		logger.GetLogger().Fatalf("init redis failed, %v", err)
	}

	// 初始化 mongo.
	if err := mongo.Init(cfg.DB.Mongo); err != nil {
		logger.GetLogger().Fatalf("init mongo failed, %v", err)
	}

	// 初始化数据库后回调.
	applifecycle.AfterInitDatabase()

	// 启动 Actor
	if err := startActor(); err != nil {
		logger.GetLogger().Fatalf("start actor failed, %v", err)
	}

	// 启动http服务.
	startHttp()
}

// Stop 停机.
func Stop() {
	stopHttp()
}

// Config 返回配置.
func Config() *config.Config {
	return cfg
}

// Env 返回环境变量.
func Env() env.Env {
	return env.Get()
}
