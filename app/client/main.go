package main

import (
	"github.com/godyy/ggs/app/client/internal/env"
	"github.com/godyy/ggs/app/client/internal/mode"
	"github.com/godyy/ggs/internal/libs/flags"
	"github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/ggs/internal/utils"
	"github.com/godyy/glog"
)

func main() {
	// 解析 flags
	flags.Parse()

	// 初始化env
	env.Init()

	// 初始化 logger
	logger.Init(&logger.Config{
		Level:       glog.DebugLevel,
		Caller:      true,
		Development: true,
		EnableStd:   true,
	})

	mode := mode.CreateMode(env.Mode)
	mode.Start()

	utils.ListenShutdown()
	mode.Stop()
}
