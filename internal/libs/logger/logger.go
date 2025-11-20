package logger

import (
	"github.com/godyy/glog"
)

// Config 日志配置
type Config struct {
	// Level 日志等级
	Level glog.Level

	// Caller 是否记录调用者
	Caller bool

	// CallerSkip 调用者跳过层数
	CallerSkip int

	// Development 是否开发模式
	Development bool

	// EnableStd 是否启用标准输出
	EnableStd bool

	// FileParams 日志文件输出相关 Core 配置参数
	FileParams *glog.FileCoreParams
}

type Logger = glog.Logger

// logger 日志实例.
var logger glog.Logger

// Init 初始化日志.
func Init(cfg *Config) {
	glogCfg := &glog.Config{
		Level:        cfg.Level,
		EnableCaller: cfg.Caller,
		CallerSkip:   cfg.CallerSkip,
		Development:  cfg.Development,
	}
	if cfg.EnableStd {
		glogCfg.Cores = append(glogCfg.Cores, glog.NewStdCoreConfig())
	}
	if cfg.FileParams != nil {
		glogCfg.Cores = append(glogCfg.Cores, glog.NewFileCoreConfig(cfg.FileParams))
	}
	logger = glog.NewLogger(glogCfg)
	invokeAfterInitFuncs()
}

// GetLogger 获取日志实例.
func GetLogger() glog.Logger {
	return logger
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
		f(logger)
	}
}
