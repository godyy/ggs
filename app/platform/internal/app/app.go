package app

import (
	"context"
	"net/http"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/platform/internal/base/config"
	"github.com/godyy/ggs/app/platform/internal/base/env"
	"github.com/godyy/ggs/app/platform/internal/infra/repo"
	applifecycle "github.com/godyy/ggs/internal/base/lifecycle"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	mongomodels "github.com/godyy/ggs/internal/infra/mongo/models"
	"github.com/godyy/ggskit/base/db/mongo"
	"github.com/godyy/ggskit/base/db/redis"
	"github.com/godyy/ggskit/base/flags"
	"github.com/godyy/ggskit/infra/cluster"
	pkgerrors "github.com/pkg/errors"
)

type app struct {
	cfg *config.Config // 配置
	env *env.Env       // 环境变量管理器.

	redisClient redis.Client  // redis 客户端
	mongoClient *mongo.Client // mongo 客户端

	actorRegistry    gactor.ActorRegistry
	actorRouter      *actor.Router
	actorService     *actor.Service
	actorCodec       *actor.Codec
	actorServerStore *actor.ServerStore

	cluster *cluster.Service // cluster.

	httpServer *http.Server // http 服务
}

var appInst *app

// Start 启动.
func Start() {
	appInst = &app{}

	// 解析flag
	flags.Parse()
	defer flags.Reset()

	// 加载配置.
	if c, err := config.Load(configPath()); err != nil {
		panic(pkgerrors.WithMessage(err, "load config"))
	} else {
		appInst.cfg = c
	}

	// 初始化环境变量.
	appInst.env = env.NewEnv()
	appInst.env.Init()

	// 初始化日志工具.
	logger.Init(appInst.cfg.Log)

	// 启动前回调.
	applifecycle.BeforeStart()

	// 初始化 redis.
	redisCli, err := redis.NewClient(appInst.cfg.DB.Redis)
	if err != nil {
		logger.Get().Fatalf("init redis failed, %v", err)
	}
	appInst.redisClient = redisCli

	// 初始化 mongo.
	cli, err := mongo.Connect(appInst.cfg.DB.Mongo)
	if err != nil {
		logger.Get().Fatalf("init mongo failed, %v", err)
	}
	appInst.mongoClient = cli
	if err := ensureMongoIndexes(); err != nil {
		logger.Get().Fatalf("ensure mongo indexes failed, %v", err)
	}

	// 初始化 repo.
	repo.Init(appInst.mongoClient)

	// 启动 Actor
	if err := appInst.startActor(); err != nil {
		logger.Get().Fatalf("start actor failed, %v", err)
	}

	// 启动 Actor 服务.
	if err := appInst.startActor(); err != nil {
		logger.Get().Fatalf("start actor failed, %v", err)
	}

	// 启动 cluster.
	if err := appInst.startCluster(); err != nil {
		logger.Get().Fatalf("start cluster failed, %v", err)
	}

	// 启动http服务.
	appInst.startHttp()
}

// Stop 停机.
func Stop() {
	// 停止 http 服务.
	appInst.stopHttp()

	// 停止 Actor 服务.
	appInst.stopActor()

	if appInst.mongoClient != nil {
		if err := appInst.mongoClient.Disconnect(context.Background()); err != nil {
			logger.Get().Errorf("disconnect mongo failed, %v", err)
		}
	}
	if appInst.redisClient != nil {
		if err := appInst.redisClient.Close(); err != nil {
			logger.Get().Errorf("close redis failed, %v", err)
		}
	}

	// 停止 cluster.
	appInst.stopCluster()
}

// Config 返回配置.
func Config() *config.Config {
	return appInst.cfg
}

// Env 返回环境变量.
func Env() *env.Env {
	return appInst.env
}

// ensureMongoIndexes 确保 mongo 索引存在.
func ensureMongoIndexes() error {
	return mongomodels.EnsureIndexes(context.Background(), appInst.mongoClient, mongomodels.DBPlatform,
		mongomodels.CollAccount, mongomodels.CollCharacter, mongomodels.CollServer, mongomodels.CollIDGenerator)
}
