package app

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/base/config"
	"github.com/godyy/ggs/app/game/internal/base/env"
	applifecycle "github.com/godyy/ggs/internal/base/lifecycle"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/gdconf"
	imongobd "github.com/godyy/ggs/internal/infra/mongobd"
	"github.com/godyy/ggskit/base/db/mongo"
	"github.com/godyy/ggskit/base/db/redis"
	"github.com/godyy/ggskit/base/flags"
	"github.com/godyy/ggskit/infra/actor"
	"github.com/godyy/ggskit/infra/cluster"
	"github.com/godyy/ggskit/infra/mongobd"
	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
)

type app struct {
	config *config.Config // 配置

	env *env.Env // 环境变量管理器.

	redisClient redis.Client  // redis 客户端.
	mongoClient *mongo.Client // mongo 客户端.
	mongobd     *imongobd.BD  // mongo 后台.

	cluster *cluster.Service // cluster.

	actorProtoReg    *actor.ProtoRegistry // actor 协议注册表
	actorCodec       *actor.Codec         // actor编解码
	actorRegistry    gactor.ActorRegistry // actor 注册表
	actorServerStore *actor.ServerStore   // actor 所属服务器存储
	actorRouter      *actor.Router        // actor 节点路由
	actorService     *actor.Service       // actor 服务

	httpServer *http.Server // http 服务
}

var appInst *app

func Start() {
	appInst = &app{}

	// 解析flag
	flags.Parse()
	defer flags.Reset()

	// 加载配置表
	cfg, err := config.Load(configPath())
	if err != nil {
		panic(pkgerrors.WithMessage(err, "load config"))
	}
	appInst.config = cfg

	// 初始化环境变量.
	appInst.env = env.NewEnv()
	appInst.env.Init()

	// 初始化日志工具.
	logger.Init(cfg.Log)

	// 启动前回调.
	applifecycle.BeforeStart()

	// 初始化 redis.
	redisClient, err := redis.NewClient(cfg.DB.Redis)
	if err != nil {
		logger.Get().Fatalf("init redis failed, %v", err)
	}
	appInst.redisClient = redisClient

	// 初始化 mongo.
	mongoClient, mongoErr := mongo.Connect(cfg.DB.Mongo)
	if mongoErr != nil {
		logger.Get().Fatalf("init mongo failed, %v", mongoErr)
	}
	appInst.mongoClient = mongoClient

	// 启动mongo后台.
	mongobd, err := imongobd.New(imongobd.Config{
		BDConfig: mongobd.BDConfig{
			Client:         appInst.mongoClient,
			Wokers:         runtime.NumCPU(),
			MaxWorkerOps:   1000,
			DefExecTimeout: time.Second * 5,
			Logger:         logger.Get(),
		},
		OpChanSize:  10000,
		OpConsumers: 2,
	})
	if err != nil {
		logger.Get().Fatalf("start mongobd failed, %v", err)
	}
	appInst.mongobd = mongobd

	// 配置表加载.
	gdconfDB := mongoClient.Database("gdconf", options.Database().SetReadConcern(readconcern.Majority()))
	if err := gdconf.Load(gdconfDB); err != nil {
		logger.Get().Fatalf("load gdconf failed, %v", err)
	}
	logger.Get().Info("load gdconf success")

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

func Stop() {
	// 停止 http 服务.
	appInst.stopHttp()

	// 停止 Actor 服务.
	appInst.stopActor()

	// 停止 mongo 后台.
	appInst.mongobd.Stop()
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
	return appInst.config
}

// Env 返回环境变量管理器.
func Env() *env.Env {
	return appInst.env
}

// MongoBD 返回 mongo 后台.
func MongoBD() *imongobd.BD {
	return appInst.mongobd
}
