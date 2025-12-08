package app

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/ggs/internal/libs/pprof"
)

func (a *app) startHttp() {
	port := a.config.HttpPort
	if port <= 0 || !a.config.EnablePProf {
		return
	}

	mux := http.NewServeMux()
	if a.config.EnablePProf {
		pprof.RegisterHTTP(mux, "")
	}

	a.httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	go func() {
		logger.GetLogger().Infof("http server listening at :%d", port)
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
		ctx, cancel := context.WithTimeout(context.Background(), consts.ShutdownTimeout)
		defer cancel()

		// 优雅关闭服务器
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.GetLogger().Error("http server shutdown with error: %v", err)
		}
	}
}
