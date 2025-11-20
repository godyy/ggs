package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/godyy/ggs/app/login/internal/config"
	mactor "github.com/godyy/ggs/app/login/internal/modules/actor"
	"github.com/godyy/ggs/internal/libs/logger"
	pkgerrors "github.com/pkg/errors"
)

// shutdownTimeout 停机超时.
const shutdonwTimeout = 30 * time.Second

var (
	srv *http.Server
)

// Start 启动
func Start() error {
	if err := mactor.Init(); err != nil {
		return pkgerrors.WithMessage(err, "init actor module")
	}

	// 启动http服务.
	startHttp()

	return nil
}

// Stop 停机.
func Stop() {
	if srv != nil {
		// 创建上下文用于优雅关闭
		ctx, cancel := context.WithTimeout(context.Background(), shutdonwTimeout)
		defer cancel()

		// 优雅关闭服务器
		if err := srv.Shutdown(ctx); err != nil {
			logger.GetLogger().Error("http server shutdown with error: %v", err)
		}
	}
}

func startHttp() {
	// 创建gin引擎
	engine := gin.New()

	// 配置路由
	setupRoutes(engine)

	// 创建http服务
	srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GetConfig().Port),
		Handler: engine,
	}

	// 启动http服务
	go func() {
		logger.GetLogger().Infof("http server listening at :%d", config.GetConfig().Port)
		if err := srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			logger.GetLogger().Info("http server closed.")
		} else {
			logger.GetLogger().Errorf("http server closed with error: %v", err)
		}
	}()
}
