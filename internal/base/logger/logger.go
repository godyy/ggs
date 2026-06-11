package logger

import (
	"sync"

	"github.com/godyy/ggskit/base/logger"
)

type Logger = logger.Logger

type Config = logger.Config

// _logger 日志实例.
var _logger Logger

// once 日志初始化一次.
var once sync.Once

// Init 初始化
func Init(cfg *Config) {
	once.Do(func() {
		_logger = logger.CreateLogger(cfg)
		invokeAfterInitFuncs()
	})
}

// Get 获取日志实例.
func Get() Logger {
	return _logger
}

// AfterInitFunc 日志初始化后回调函数.
type AfterInitFunc func(logger Logger)

// afterInitFuncs 日志初始化后回调函数列表.
var afterInitFuncs []AfterInitFunc

// RegisterAfterInitFunc 注册日志初始化后回调函数.
func RegisterAfterInitFunc(f AfterInitFunc) {
	afterInitFuncs = append(afterInitFuncs, f)
}

// invokeAfterInitFuncs 日志初始化后回调.
func invokeAfterInitFuncs() {
	for _, f := range afterInitFuncs {
		f(_logger)
	}
}
