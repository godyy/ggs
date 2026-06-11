package app

import (
	"net"
	"net/http"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/agent/internal"
	"github.com/godyy/ggs/app/agent/internal/base/config"
	"github.com/godyy/ggs/app/agent/internal/base/env"
	"github.com/godyy/ggs/app/agent/internal/infra/router"
	icrypto "github.com/godyy/ggs/internal/base/crypto"
	applifecycle "github.com/godyy/ggs/internal/base/lifecycle"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggskit/base/crypto"
	"github.com/godyy/ggskit/base/db/redis"
	"github.com/godyy/ggskit/base/flags"
	"github.com/godyy/ggskit/infra/actor"
	"github.com/godyy/ggskit/infra/cluster"
	"github.com/godyy/ggskit/infra/noderouter"
	pkgerrors "github.com/pkg/errors"
)

type app struct {
	config *config.Config // 配置

	env *env.Env // 环境变量管理器

	redisClient redis.Client // redis 客户端

	// 对接 c 端
	listener net.Listener

	// cluster.
	cluster *cluster.Service

	// nodeSelector.
	nodeSelector *router.NodeSelector

	// actor.
	actorRegistry gactor.ActorRegistry
	actorClient   *actor.Client

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

	// 初始化加密工具
	if err := appInst.initCrypto(); err != nil {
		logger.Get().Fatalf("init crypto failed, %v", err)
	}

	// 启动 Actor 服务.
	if err := appInst.startActor(); err != nil {
		logger.Get().Fatalf("start actor failed, %v", err)
	}

	// 初始化节点路由选择器.
	appInst.nodeSelector = router.NewNodeSelector(noderouter.NewRendezvousSelector())

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

	if appInst.redisClient != nil {
		if err := appInst.redisClient.Close(); err != nil {
			logger.Get().Errorf("close redis failed, %v", err)
		}
	}
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
func Env() *env.Env {
	return appInst.env
}

// NodeSelector 获取节点路由选择器.
func NodeSelector() *router.NodeSelector {
	return appInst.nodeSelector
}

// ActorRegistry 获取 Actor 注册表.
func ActorRegistry() gactor.ActorRegistry {
	return appInst.actorRegistry
}

// ActorClient 获取 Actor 客户端.
func ActorClient() *actor.Client {
	return appInst.actorClient
}

// RedisClient 获取 redis 客户端.
func RedisClient() redis.Client {
	return appInst.redisClient
}
