package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/godyy/ggs/app/platform/internal"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
)

var (
	srv *http.Server
)

func startHttp() {
	// 创建gin引擎
	engine := gin.New()

	// 配置路由
	if internal.SetupRoutes == nil {
		logger.Get().Fatal("internal.SetupRoutes is nil")
	}
	internal.SetupRoutes(engine.Group("/api"))

	// 创建http服务
	srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: engine,
	}

	// 启动http服务
	go func() {
		logger.Get().Infof("http server listening at :%d", cfg.Port)
		if err := srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			logger.Get().Info("http server closed.")
		} else {
			logger.Get().Errorf("http server closed with error: %v", err)
		}
	}()
}

func stopHttp() {
	if srv != nil {
		// 创建上下文用于优雅关闭
		ctx, cancel := context.WithTimeout(context.Background(), consts.ShutdownTimeout)
		defer cancel()

		// 优雅关闭服务器
		if err := srv.Shutdown(ctx); err != nil {
			logger.Get().Error("http server shutdown with error: %v", err)
		}
	}
}
