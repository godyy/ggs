package app

import (
	"net"

	"github.com/godyy/ggs/app/agent/internal"
	"github.com/godyy/ggs/app/agent/internal/base/config"
	icrypto "github.com/godyy/ggs/app/internal/base/crypto"
	applifecycle "github.com/godyy/ggs/app/internal/base/lifecycle"
	"github.com/godyy/ggs/internal/base/crypto"
	"github.com/godyy/ggs/internal/base/env"
	"github.com/godyy/ggs/internal/libs/db/mongo"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/flags"
	"github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/ggs/internal/modules/actor"
	"github.com/godyy/ggs/internal/modules/cluster"
	pkgerrors "github.com/pkg/errors"
)

type app struct {
	config *config.Config // 配置

	// 对接 c 端
	listener net.Listener

	// cluster.
	cluster *cluster.Service

	// actor.
	actorMetaDriver *actor.MetaDriver
	actorClient     *actor.Client

	// crypto.
	secretDecryptor crypto.Decryptor
}

var appInst *app

// Start 启动应用.
func Start() {
	appInst = &app{}

	// 解析flag
	flags.Parse()
	defer flags.Clear()

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
	if err := redis.Init(cfg.DB.Redis); err != nil {
		logger.GetLogger().Fatalf("init redis failed, %v", err)
	}

	// 初始化 mongo.
	if err := mongo.Init(cfg.DB.Mongo); err != nil {
		logger.GetLogger().Fatalf("init mongo failed, %v", err)
	}

	// 初始化加密工具
	if err := appInst.initCrypto(); err != nil {
		logger.GetLogger().Fatalf("init crypto failed, %v", err)
	}

	// 启动 Actor 服务.
	if err := appInst.startActor(); err != nil {
		logger.GetLogger().Fatalf("start actor failed, %v", err)
	}

	// 启动 cluster.
	if err := appInst.startCluster(); err != nil {
		logger.GetLogger().Fatalf("start cluster failed, %v", err)
	}

	// 启动对 c 端监听服务.
	if err := appInst.startListen(); err != nil {
		logger.GetLogger().Fatalf("start listening failed, %v", err)
	}
}

// Stop 停止应用.
func Stop() {
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

// ActorClient 获取 Actor 客户端.
func ActorClient() *actor.Client {
	return appInst.actorClient
}
