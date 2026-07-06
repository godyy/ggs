package glog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level 日志等级.
type Level = zapcore.Level

const (
	DebugLevel  = zap.DebugLevel
	InfoLevel   = zap.InfoLevel
	WarnLevel   = zap.WarnLevel
	ErrorLevel  = zap.ErrorLevel
	DPanicLevel = zap.DPanicLevel
	PanicLevel  = zap.PanicLevel
	FatalLevel  = zap.FatalLevel
)

// Logger 封装日志接口.
type Logger interface {
	// ZapLogger 获取并返回内部的zap.Logger
	ZapLogger() *zap.Logger

	// Level 获取启动的最小日志level
	Level() Level

	// Name 获取 logger name
	Name() string

	// Named 为 logger 追加名称
	Named(name string) Logger

	// AddCallerSkip 增加 caller skip
	AddCallerSkip(skip int) Logger

	// WithFields 基于结构化字段创建一个 child Logger
	// 所返回的 child Logger 的每一次日志操作，都会同时输出这些结构化字段
	WithFields(fields ...zap.Field) Logger

	// WithFieldsLazy 基于结构化字段创建一个 child Logger
	// 所返回的 child Logger 的每一次日志操作，都会同时输出这些结构化字段。
	// 同WithFields的区别在于，其只会在真正用到fields的时候才会评估其值。
	WithFieldsLazy(fields ...zap.Field) Logger

	// WithOptions 复制logger并应用选项
	WithOptions(opts ...zap.Option) Logger

	// Debug for DebugLevel
	Debug(v ...any)
	Debugln(v ...any)
	Debugf(f string, v ...any)
	DebugFields(msg string, fields ...zap.Field)
	Debugw(msg string, keyAndValues ...any)

	// Info for InfoLevel
	Info(v ...any)
	Infoln(v ...any)
	Infof(f string, v ...any)
	InfoFields(msg string, fields ...zap.Field)
	Infow(msg string, keyAndValues ...any)

	// Warn for WarnLevel
	Warn(v ...any)
	Warnln(v ...any)
	Warnf(f string, v ...any)
	WarnFields(msg string, fields ...zap.Field)
	Warnw(msg string, keyAndValues ...any)

	// Error for ErrorLevel
	Error(v ...any)
	Errorln(v ...any)
	Errorf(f string, v ...any)
	ErrorFields(msg string, fields ...zap.Field)
	Errorw(msg string, keyAndValues ...any)

	// DPanic for DPanicLevel
	DPanic(v ...any)
	DPanicln(v ...any)
	DPanicf(f string, v ...any)
	DPanicFields(msg string, fields ...zap.Field)
	DPanicw(msg string, keyAndValues ...any)

	// Panic for PanicLevel
	Panic(v ...any)
	Panicln(v ...any)
	Panicf(f string, v ...any)
	PanicFields(msg string, fields ...zap.Field)
	Panicw(msg string, keyAndValues ...any)

	// Fatal for FatalLevel
	Fatal(v ...any)
	Fatalln(v ...any)
	Fatalf(f string, v ...any)
	FatalFields(msg string, fields ...zap.Field)
	Fatalw(msg string, keyAndValues ...any)
}

// Config 提供创建Logger需要用到的相关配置项.
type Config struct {
	// Level 日志等级.
	Level Level

	// EnableCaller 是否记录日志caller.
	EnableCaller bool

	// CallerSkip 默认为0.
	CallerSkip int

	// Development 是否开发模式, 控制 DPanicLevel 是否 log panic
	Development bool

	// Cores 日志核心配置.
	Cores []CoreConfig
}

// logger Logger 内部实现.
type logger struct {
	zap   *zap.Logger        // 结构化日志
	sugar *zap.SugaredLogger // printf-like 日志
}

// NewLogger 根据配置构造 Logger.
func NewLogger(c *Config) Logger {
	if c == nil {
		panic("glog: NewLogger: config is nil")
	}

	if len(c.Cores) == 0 {
		panic("glog: NewLogger: Cores not specified")
	}

	// 选项.
	var options []zap.Option
	if c.EnableCaller {
		options = append(options, zap.AddCaller())
		options = append(options, zap.AddCallerSkip(1+c.CallerSkip))
	}
	if c.Development {
		options = append(options, zap.Development())
	}

	// cores.
	cores := make([]zapcore.Core, 0, len(c.Cores))
	for _, core := range c.Cores {
		cores = append(cores, core(c))
	}

	zapLogger := zap.New(zapcore.NewTee(cores...), options...)
	return &logger{
		zap:   zapLogger,
		sugar: zapLogger.Sugar(),
	}
}

func (l *logger) ZapLogger() *zap.Logger {
	return l.zap
}

func (l *logger) Level() Level {
	return l.zap.Level()
}

func (l *logger) Name() string {
	l.zap.WithOptions()
	return l.zap.Name()
}

func (l *logger) Named(name string) Logger {
	clone := l.clone()
	clone.zap = l.zap.Named(name)
	clone.sugar = clone.zap.Sugar()
	return clone
}

func (l *logger) AddCallerSkip(skip int) Logger {
	clone := l.clone()
	clone.zap = l.zap.WithOptions(zap.AddCallerSkip(skip))
	clone.sugar = clone.zap.Sugar()
	return clone
}

func (l *logger) WithFields(fields ...zap.Field) Logger {
	clone := l.clone()
	clone.zap = l.zap.With(fields...)
	clone.sugar = clone.zap.Sugar()
	return clone
}

func (l *logger) WithFieldsLazy(fields ...zap.Field) Logger {
	clone := l.clone()
	clone.zap = l.zap.WithLazy(fields...)
	clone.sugar = clone.zap.Sugar()
	return clone
}

func (l *logger) WithOptions(opts ...zap.Option) Logger {
	clone := l.clone()
	clone.zap = l.zap.WithOptions(opts...)
	clone.sugar = clone.zap.Sugar()
	return clone
}

func (l *logger) clone() *logger {
	clone := *l
	return &clone
}

func (l *logger) Debug(v ...any) {
	l.sugar.Debug(v...)
}

func (l *logger) Debugln(v ...any) {
	l.sugar.Debugln(v...)
}

func (l *logger) Debugf(f string, v ...any) {
	l.sugar.Debugf(f, v...)
}

func (l *logger) DebugFields(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

func (l *logger) Debugw(msg string, keyAndValues ...any) {
	l.sugar.Debugw(msg, keyAndValues...)
}

func (l *logger) Info(v ...any) {
	l.sugar.Info(v...)
}

func (l *logger) Infoln(v ...any) {
	l.sugar.Infoln(v...)
}

func (l *logger) Infof(f string, v ...any) {
	l.sugar.Infof(f, v...)
}

func (l *logger) InfoFields(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

func (l *logger) Infow(msg string, keyAndValues ...any) {
	l.sugar.Infow(msg, keyAndValues...)
}

func (l *logger) Warn(v ...any) {
	l.sugar.Warn(v...)
}

func (l *logger) Warnln(v ...any) {
	l.sugar.Warnln(v...)
}

func (l *logger) Warnf(f string, v ...any) {
	l.sugar.Warnf(f, v...)
}

func (l *logger) WarnFields(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

func (l *logger) Warnw(msg string, keyAndValues ...any) {
	l.sugar.Warnw(msg, keyAndValues...)
}

func (l *logger) Error(v ...any) {
	l.sugar.Error(v...)
}

func (l *logger) Errorln(v ...any) {
	l.sugar.Errorln(v...)
}

func (l *logger) Errorf(f string, v ...any) {
	l.sugar.Errorf(f, v...)
}

func (l *logger) ErrorFields(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

func (l *logger) Errorw(msg string, keyAndValues ...any) {
	l.sugar.Errorw(msg, keyAndValues...)
}

func (l *logger) DPanic(v ...any) {
	l.sugar.DPanic(v...)
}

func (l *logger) DPanicln(v ...any) {
	l.sugar.DPanicln(v...)
}

func (l *logger) DPanicf(f string, v ...any) {
	l.sugar.DPanicf(f, v...)
}

func (l *logger) DPanicFields(msg string, fields ...zap.Field) {
	l.zap.DPanic(msg, fields...)
}

func (l *logger) DPanicw(msg string, keyAndValues ...any) {
	l.sugar.DPanicw(msg, keyAndValues...)
}

func (l *logger) Panic(v ...any) {
	l.sugar.Panic(v...)
}

func (l *logger) Panicln(v ...any) {
	l.sugar.Panicln(v...)
}

func (l *logger) Panicf(f string, v ...any) {
	l.sugar.Panicf(f, v...)
}

func (l *logger) PanicFields(msg string, fields ...zap.Field) {
	l.zap.Panic(msg, fields...)
}

func (l *logger) Panicw(msg string, keyAndValues ...any) {
	l.sugar.Panicw(msg, keyAndValues...)
}

func (l *logger) Fatal(v ...any) {
	l.sugar.Fatal(v...)
}

func (l *logger) Fatalln(v ...any) {
	l.sugar.Fatalln(v...)
}

func (l *logger) Fatalf(f string, v ...any) {
	l.sugar.Fatalf(f, v...)
}

func (l *logger) FatalFields(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

func (l *logger) Fatalw(msg string, keyAndValues ...any) {
	l.sugar.Fatalw(msg, keyAndValues...)
}
