package app

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/godyy/ggs/internal/libs/db/mongo"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/ggs/internal/libs/probe"
)

// shutdownTimeout 停机超时.
const shutdonwTimeout = 30 * time.Second

func (a *app) startHttp() {
	if a.config.HttpPort <= 0 {
		return
	}

	mux := http.NewServeMux()
	a.registerHttpProbe(mux)

	a.httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(a.config.HttpPort),
		Handler: mux,
	}

	go func() {
		logger.GetLogger().Infof("http server listening at :%d", a.config.HttpPort)
		if err := a.httpServer.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			logger.GetLogger().Info("http server closed.")
		} else {
			logger.GetLogger().Errorf("http server closed with error: %v", err)
		}
	}()
}

func (a *app) stopHttp() {
	if a.httpServer != nil {
		// 创建上下文用于优雅关闭
		ctx, cancel := context.WithTimeout(context.Background(), shutdonwTimeout)
		defer cancel()

		// 优雅关闭服务器
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.GetLogger().Error("http server shutdown with error: %v", err)
		}
	}
}

func (a *app) registerHttpProbe(mux *http.ServeMux) {
	probe.Init(
		probe.WithReadinessPolicy(probe.Cached),
		probe.WithReadinessCacheTTL(5*time.Second),
		probe.WithReadinessTimeout(5*time.Second),
		probe.WithReadinessChecker("mongo", func(ctx context.Context) error {
			return mongo.Inst().Ping(ctx, nil)
		}),
		probe.WithReadinessChecker("redis", func(ctx context.Context) error {
			return redis.Inst().Ping(ctx).Err()
		}),
	)

	probe.SetReady(true)
	probe.RegisterHTTP(mux, "/healthz", "/readyz")
}
