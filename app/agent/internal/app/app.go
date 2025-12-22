package app

import (
	"net"
	"net/http"

	"github.com/godyy/ggs/app/agent/internal"
	"github.com/godyy/ggs/app/agent/internal/base/config"
	"github.com/godyy/ggs/app/agent/internal/infra/router"
	icrypto "github.com/godyy/ggs/app/internal/base/crypto"
	applifecycle "github.com/godyy/ggs/app/internal/base/lifecycle"
	"github.com/godyy/ggs/internal/base/crypto"
	mongocli "github.com/godyy/ggs/internal/base/db/mongo/cli"
	rediscli "github.com/godyy/ggs/internal/base/db/redis/cli"
	"github.com/godyy/ggs/internal/base/env"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/cluster"
	baserouter "github.com/godyy/ggs/internal/infra/router"
	"github.com/godyy/gutils/flags"
	pkgerrors "github.com/pkg/errors"
)

type app struct {
	config *config.Config // 配置

	// 对接 c 端
	listener net.Listener

	// cluster.
	cluster *cluster.Service

	// nodeSelector.
	nodeSelector *router.NodeSelector

	// actor.
	actorMetaDriver *actor.MetaDriver
	actorClient     *actor.Client

	// crypto.
	secretDecryptor crypto.Decryptor

	httpServer *http.Server // http 服务
}

var appInst *app

// Start 启动应用.
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

	// 初始化加密工具
	if err := appInst.initCrypto(); err != nil {
		logger.Get().Fatalf("init crypto failed, %v", err)
	}

	// 启动 Actor 服务.
	if err := appInst.startActor(); err != nil {
		logger.Get().Fatalf("start actor failed, %v", err)
	}

	// 初始化节点路由选择器.
	appInst.nodeSelector = router.NewNodeSelector(baserouter.NewRendezvousSelector())

	// 启动 cluster.
	if err := appInst.startCluster(); err != nil {
		logger.Get().Fatalf("start cluster failed, %v", err)
	}

	// 启动对 c 端监听服务.
	if err := appInst.startListen(); err != nil {
		logger.Get().Fatalf("start listening failed, %v", err)
	}

	// 启动 http 服务.
	appInst.startHttp()
}

// Stop 停止应用.
func Stop() {
	// 停止 http 服务.
	appInst.stopHttp()

	// 停止对 c 端监听服务.
	appInst.stopListen()

	// 停止所有 agent.
	appInst.stopAllAgents()

	// 停止 Actor 服务.
	appInst.stopActor()

	// 停止 cluster.
	appInst.stopCluster()
}

// initCrypto 初始化加密工具.
func (a *app) initCrypto() error {
	if secretDecryptor, err := icrypto.CreateRSADecryptor(); err != nil {
		return pkgerrors.WithMessage(err, "create secret decryptor")
	} else {
		a.secretDecryptor = secretDecryptor
	}
	return nil
}

// stopAllAgents 停止所有 agent.
func (a *app) stopAllAgents() {
	internal.StopAllAgents()
}

// Config 获取配置.
func Config() *config.Config {
	return appInst.config
}

// Env 获取环境变量.
func Env() env.Env {
	return env.Get()
}

// NodeSelector 获取节点路由选择器.
func NodeSelector() *router.NodeSelector {
	return appInst.nodeSelector
}

// ActorMetaDriver 获取 Actor Meta 驱动.
func ActorMetaDriver() *actor.MetaDriver {
	return appInst.actorMetaDriver
}

// ActorClient 获取 Actor 客户端.
func ActorClient() *actor.Client {
	return appInst.actorClient
}
