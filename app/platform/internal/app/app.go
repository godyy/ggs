package app

import (
	"context"

	"github.com/godyy/ggs/app/platform/internal/base/config"
	"github.com/godyy/ggs/app/platform/internal/infra/repo"
	applifecycle "github.com/godyy/ggs/internal/base/lifecycle"
	"github.com/godyy/ggs/internal/base/logger"
	mongomodels "github.com/godyy/ggs/internal/infra/mongo/models"
	"github.com/godyy/ggskit/base/db/mongo"
	"github.com/godyy/ggskit/base/db/redis"
	"github.com/godyy/ggskit/base/env"
	"github.com/godyy/ggskit/base/flags"
	pkgerrors "github.com/pkg/errors"
)

var (
	cfg         *config.Config // 配置
	redisClient redis.Client   // redis 客户端
	mongoClient *mongo.Client  // mongo 客户端
)

// Start 启动.
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
	redisCli, err := redis.NewClient(cfg.DB.Redis)
	if err != nil {
		logger.Get().Fatalf("init redis failed, %v", err)
	}
	redisClient = redisCli

	// 初始化 mongo.
	cli, err := mongo.Connect(cfg.DB.Mongo)
	if err != nil {
		logger.Get().Fatalf("init mongo failed, %v", err)
	}
	mongoClient = cli
	if err := ensureMongoIndexes(); err != nil {
		logger.Get().Fatalf("ensure mongo indexes failed, %v", err)
	}

	// 初始化 repo.
	repo.Init(mongoClient)

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
	if mongoClient != nil {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			logger.Get().Errorf("disconnect mongo failed, %v", err)
		}
	}
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Get().Errorf("close redis failed, %v", err)
		}
	}
}

// Config 返回配置.
func Config() *config.Config {
	return cfg
}

// Env 返回环境变量.
func Env() env.Env {
	return env.Get()
}

// ensureMongoIndexes 确保 mongo 索引存在.
func ensureMongoIndexes() error {
	return mongomodels.EnsureIndexes(context.Background(), mongoClient, mongomodels.DBPlatform,
		mongomodels.CollAccount, mongomodels.CollCharacter, mongomodels.CollServer, mongomodels.CollIDGenerator)
}
