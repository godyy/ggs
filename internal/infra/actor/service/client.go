package service

import (
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/logger"
	iactor "github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggskit/base/env"
	"github.com/godyy/ggskit/infra/actor"
)

// ClientConfig 客户端配置.
type ClientConfig struct {
	NodeId  string               // 节点 ID.
	Handler gactor.ClientHandler // 处理器.
}

// NewClient 创建 actor 客户端.
func NewClient(cfg *ClientConfig) *iactor.Client {
	// 创建 actor 客户端.
	clientCfg := &actor.ClientConfig{
		Core: &gactor.ClientConfig{
			NodeId:            cfg.NodeId,
			ActorCategory:     iactor.CategoryPlayer.ActorCategory(),
			DefRequestTimeout: time.Second * 10,
			Handler:           cfg.Handler,
		},
		Logger: logger.Get(),
	}
	if env.Get().Debug() {
		clientCfg.Core.DefCtxTimeout = time.Hour * 1
	}
	return actor.NewClient(clientCfg)
}
