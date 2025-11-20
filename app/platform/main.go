package main

import (
	"github.com/godyy/ggs/app/platform/internal/app"
	"github.com/godyy/ggs/app/platform/internal/config"
	"github.com/godyy/ggs/internal/env"
	"github.com/godyy/ggs/internal/libs/db/mongo"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/flags"
	"github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/ggs/internal/utils"
	pkgerrors "github.com/pkg/errors"
)

func main() {
	configPath := flags.String("config-path", "./configs/dev.toml", "config path")
	flags.Parse()

	// 初始化配置.
	if err := config.Init(*configPath); err != nil {
		panic(pkgerrors.WithMessage(err, "init config"))
	}

	// 初始化日志工具.
	logger.Init(config.GetConfig().Log)

	// 初始化env
	env.Init()

	// 初始化 redis.
	if err := redis.Init(config.GetConfig().DB.Redis); err != nil {
		logger.GetLogger().Fatalf("init redis failed, %v", err)
	}

	// 初始化 mongo.
	if err := mongo.Init(config.GetConfig().DB.Mongo); err != nil {
		logger.GetLogger().Fatalf("init mongo failed, %v", err)
	}

	// 启动.
	if err := app.Start(); err != nil {
		logger.GetLogger().Fatalf("start app failed, %v", err)
	}

	// 监听关闭信号.
	utils.ListenShutdown()

	// 停机.
	app.Stop()
}
