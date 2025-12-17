package app

import (
	"net/http"
	"runtime"
	"time"

	"github.com/godyy/ggs/app/game/internal/base/config"
	"github.com/godyy/ggs/app/game/internal/base/env"
	applifecycle "github.com/godyy/ggs/app/internal/base/lifecycle"
	imongobd "github.com/godyy/ggs/app/internal/infra/mongobd"
	mongocli "github.com/godyy/ggs/internal/base/db/mongo/cli"
	rediscli "github.com/godyy/ggs/internal/base/db/redis/cli"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/cluster"
	"github.com/godyy/ggs/internal/infra/mongobd"
	"github.com/godyy/gutils/flags"
	pkgerrors "github.com/pkg/errors"
)

type app struct {
	config *config.Config // 配置

	env *env.Env // 环境变量管理器.

	mongobd *imongobd.BD // mongo 后台.

	cluster *cluster.Service // cluster.

	actorCodec      actor.Codec       // actor编解码
	actorMetaDriver *actor.MetaDriver // actor Meta驱动
	actorService    *actor.Service    // actor 服务

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
	if err := rediscli.Init(cfg.DB.Redis); err != nil {
		logger.Get().Fatalf("init redis failed, %v", err)
	}

	// 初始化 mongo.
	if err := mongocli.Init(cfg.DB.Mongo); err != nil {
		logger.Get().Fatalf("init mongo failed, %v", err)
	}

	// 启动mongo后台.
	mongobd, err := imongobd.New(imongobd.Config{
		BDConfig: mongobd.BDConfig{
			Client:         mongocli.Get(),
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
