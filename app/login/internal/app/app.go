package app

import (
	applifecycle "github.com/godyy/ggs/app/internal/base/lifecycle"
	"github.com/godyy/ggs/app/login/internal/base/config"
	dbrepo "github.com/godyy/ggs/app/login/internal/base/db/repo"
	mongocli "github.com/godyy/ggs/internal/base/db/mongo/cli"
	rediscli "github.com/godyy/ggs/internal/base/db/redis/cli"
	"github.com/godyy/ggs/internal/base/env"
	_ "github.com/godyy/ggs/internal/base/env"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/gutils/flags"
	pkgerrors "github.com/pkg/errors"
)

var (
	cfg *config.Config // 配置
)

// Start 启动
func Start() {
	// 解析flag
	flags.Parse()
	defer flags.Reset()

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
	if err := rediscli.Init(cfg.DB.Redis); err != nil {
		logger.Get().Fatalf("init redis failed, %v", err)
	}

	// 初始化 mongo.
	if err := mongocli.Init(cfg.DB.Mongo); err != nil {
		logger.Get().Fatalf("init mongo failed, %v", err)
	}

	// 初始化db repo.
	dbrepo.Init()

	// 启动 Actor
	if err := startActor(); err != nil {
		logger.Get().Fatalf("start actor failed, %v", err)
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

// Env 返回环境变量管理器.
func Env() env.Env {
	return env.Get()
}
